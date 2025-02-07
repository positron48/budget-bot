import argparse
import subprocess
import re
import os
import json
import csv
from datetime import datetime

import matplotlib.pyplot as plt
import pandas as pd
import numpy as np


def analyze_cursor_log(log_file_path):
    """Парсит файл cursor-log.md для извлечения промптов.
    Промпты разделены тремя дефисами ("---"). Пустые сегменты игнорируются.
    Вычисляет длину промпта в словах и символах.
    Возвращает список словарей с данными и сводную статистику."""
    if not os.path.exists(log_file_path):
        print(f"File {log_file_path} does not exist.")
        return []

    with open(log_file_path, 'r', encoding='utf-8') as f:
        content = f.read()

    # Разбиваем на промпты, используя разделитель '---'
    segments = content.split('---')
    prompts = []
    for segment in segments:
        prompt_text = segment.strip()
        if not prompt_text:
            continue
        # Если промпт начинается со слова 'commit' или содержит явные технические данные, пропускаем
        if prompt_text.lower().startswith('commit') or re.match(r'^\[.*\]', prompt_text):
            continue
        words = prompt_text.split()
        char_count = len(prompt_text)
        word_count = len(words)
        prompts.append({
            'text': prompt_text,
            'word_count': word_count,
            'char_count': char_count
        })

    # Выводим базовую статистику
    total_prompts = len(prompts)
    if total_prompts > 0:
        avg_words = sum(p['word_count'] for p in prompts) / total_prompts
        avg_chars = sum(p['char_count'] for p in prompts) / total_prompts
    else:
        avg_words = avg_chars = 0
    stats = {
        'total_prompts': total_prompts,
        'avg_words_per_prompt': avg_words,
        'avg_chars_per_prompt': avg_chars
    }

    print("Cursor Log Analysis:")
    print(json.dumps(stats, indent=4, ensure_ascii=False))

    # Сохраним данные в CSV
    csv_file = 'cursor_prompts.csv'
    with open(csv_file, 'w', encoding='utf-8', newline='') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=['text', 'word_count', 'char_count'])
        writer.writeheader()
        for p in prompts:
            writer.writerow(p)
    print(f"Prompts data saved to {csv_file}")

    return prompts


def analyze_git_log():
    """Извлекает историю коммитов с помощью git log и разбивает данные по коммитам.
    Использует формат: %H;%ct;%s для идентификатора, времени и сообщения коммита.
    Затем парсит строки numstat для строк добавленных и удалённых файлов.
    Возвращает список словарей с данными по коммитам."""
    try:
        # Запускаем команду git log с numstat
        git_log_cmd = ["git", "log", "--numstat", "--pretty=format:%H;%ct;%s"]
        output = subprocess.check_output(git_log_cmd, universal_newlines=True)
    except subprocess.CalledProcessError as e:
        print("Error running git log", e)
        return []

    commits = []
    current_commit = None
    for line in output.splitlines():
        line = line.strip()
        if not line:
            continue
        # Если строка содержит ;, возможно это заголовок коммита
        if re.match(r'^[0-9a-f]{40};', line):
            # Сохраняем предыдущий коммит
            if current_commit is not None:
                commits.append(current_commit)
            parts = line.split(';', 2)
            if len(parts) < 3:
                continue
            commit_hash, timestamp, message = parts
            current_commit = {
                'commit': commit_hash,
                'timestamp': int(timestamp),
                'datetime': datetime.fromtimestamp(int(timestamp)),
                'message': message,
                'additions': 0,
                'deletions': 0,
                'files_changed': 0,
                'php_yaml_additions': 0,
                'php_yaml_deletions': 0,
                'test_additions': 0,
                'test_deletions': 0,
                'main_additions': 0,
                'main_deletions': 0
            }
        else:
            # Строки numstat выглядят как: additions \t deletions \t filename
            m = re.match(r'^(\d+|-)[ \t]+(\d+|-)[ \t]+(.+)$', line)
            if m and current_commit is not None:
                add_str, del_str, filename = m.groups()
                additions = int(add_str) if add_str != '-' else 0
                deletions = int(del_str) if del_str != '-' else 0
                current_commit['additions'] += additions
                current_commit['deletions'] += deletions
                current_commit['files_changed'] += 1
                
                # Подсчет изменений только для PHP и YAML файлов
                if filename.endswith(('.php', '.yaml', '.yml')):
                    current_commit['php_yaml_additions'] += additions
                    current_commit['php_yaml_deletions'] += deletions
                    
                    # Отдельный подсчет для тестов и основного кода
                    if 'tests/' in filename or 'Tests/' in filename:
                        current_commit['test_additions'] = current_commit.get('test_additions', 0) + additions
                        current_commit['test_deletions'] = current_commit.get('test_deletions', 0) + deletions
                    else:
                        current_commit['main_additions'] = current_commit.get('main_additions', 0) + additions
                        current_commit['main_deletions'] = current_commit.get('main_deletions', 0) + deletions
    
    if current_commit is not None:
        commits.append(current_commit)

    # Сохраним данные в CSV
    csv_file = 'git_commits.csv'
    fieldnames = ['commit', 'timestamp', 'datetime', 'message', 'additions', 'deletions', 
                 'files_changed', 'php_yaml_additions', 'php_yaml_deletions',
                 'test_additions', 'test_deletions', 'main_additions', 'main_deletions']
    with open(csv_file, 'w', encoding='utf-8', newline='') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=fieldnames)
        writer.writeheader()
        for c in commits:
            writer.writerow(c)
    print(f"Git commit data saved to {csv_file}")

    return commits


def classify_prompts(prompts):
    """Автоматическая классификация промптов по ключевым словам с возможностью интерактивной корректировки для тех, что определены как 'other'.
    Предлагается выбрать тип из:
    1. feat
    2. fix
    3. refactor
    4. docs
    5. test
    6. ci
    7. continue
    8. git
    9. other
    Возвращает обновленный список промптов."""
    # Обновляем список допустимых типов для интерактивной классификации
    valid_types = ['feat', 'fix', 'refactor', 'docs', 'test', 'ci', 'continue', 'git', 'other']
    
    # Загружаем сохраненные классификации, если они есть
    saved_classifications = {}
    if os.path.exists('prompt_classifications.csv'):
        with open('prompt_classifications.csv', 'r', encoding='utf-8', newline='') as csvfile:
            reader = csv.DictReader(csvfile)
            for row in reader:
                saved_classifications[row['text']] = row['type']
    
    # Автоматическая классификация с учетом новых правил
    for p in prompts:
        # Если промпт уже классифицирован ранее, используем сохраненную классификацию
        if p['text'] in saved_classifications:
            p['type'] = saved_classifications[p['text']]
            continue

        text_lower = p['text'].lower()
        if re.search(r'\bci\b', text_lower) or 'phpstan' in text_lower or 'cs-fix' in text_lower or 'php-cs' in text_lower:
            p['type'] = 'ci'
        elif 'закоммить' in text_lower:
            p['type'] = 'git'
        elif 'commit' in text_lower or 'коммит' in text_lower:
            p['type'] = 'git'
        elif any(x in text_lower for x in ['неверн', 'не раб', 'не отраб', 'исправ', 'ошибка', 'фикс', 'error', 'unable', 'cannot', 'docker', 'докер', 'все еще', 'всё еще']):
            p['type'] = 'fix'
        elif 'продолжай' in text_lower:
            p['type'] = 'continue'
        elif 'давай' in text_lower:
            p['type'] = 'feat'
        elif 'добав' in text_lower:
            p['type'] = 'feat'
        elif 'рефактор' in text_lower:
            p['type'] = 'refactor'
        elif 'тест' in text_lower:
            p['type'] = 'test'
        elif 'ридми' in text_lower or 'readme' in text_lower or 'докум' in text_lower:
            p['type'] = 'docs'
        else:
            p['type'] = 'other'
    
    # Вывод предварительной классификации
    types = {}
    for p in prompts:
        types[p['type']] = types.get(p['type'], 0) + 1
    print("Предварительная классификация промптов:")
    print(json.dumps(types, indent=4, ensure_ascii=False))

    # Интерактивная корректировка для тех, что определены как 'other'
    for p in prompts:
        if p['type'] == 'other':
            print("\n========== Промпт ==========")
            print(p['text'])
            print("========== Конец Промпта ==========\n")
            print("Выберите новый тип для этого промпта:")
            for i, t in enumerate(valid_types, 1):
                print(f"{i}. {t}")
            choice = input("Введите номер (оставьте пустым для сохранения 'other'): ")
            if choice.strip():
                try:
                    idx = int(choice.strip()) - 1
                    if 0 <= idx < len(valid_types):
                        p['type'] = valid_types[idx]
                    else:
                        print("Неверный номер, оставляем 'other'.")
                except Exception as e:
                    print("Ошибка ввода, оставляем 'other'.")

    # Сохраняем все классификации в CSV
    with open('prompt_classifications.csv', 'w', encoding='utf-8', newline='') as csvfile:
        writer = csv.DictWriter(csvfile, fieldnames=['text', 'type'])
        writer.writeheader()
        for p in prompts:
            writer.writerow({'text': p['text'], 'type': p['type']})

    # Вывод обновленной классификации
    types_updated = {}
    for p in prompts:
        types_updated[p['type']] = types_updated.get(p['type'], 0) + 1
    print("Обновленная классификация промптов:")
    print(json.dumps(types_updated, indent=4, ensure_ascii=False))

    # Вывод количества оставшихся элементов с типом 'other'
    other_count = sum(1 for p in prompts if p['type'] == 'other')
    print(f"Количество оставшихся 'other': {other_count}")

    return prompts


def plot_graphs(prompts, commits, output_dir='output'):
    """Строит базовые графики и сохраняет их в указанную директорию."""
    if not os.path.exists(output_dir):
        os.makedirs(output_dir)

    # График соотношения слов и символов в промптах
    prompt_chars = [p['char_count'] for p in prompts]
    prompt_words = [p['word_count'] for p in prompts]
    
    plt.figure(figsize=(24,16))
    
    # Создаем scatter plot
    plt.scatter(prompt_words, prompt_chars, 
               color='#7CB9E8', alpha=0.6, s=200,
               edgecolor='darkblue', linewidth=2)
    
    plt.title('Соотношение количества слов и символов в промптах', fontsize=36)
    plt.xlabel('Количество слов', fontsize=28)
    plt.ylabel('Количество символов', fontsize=28)
    plt.xticks(fontsize=24)
    plt.yticks(fontsize=24)
    
    # Добавляем сетку
    plt.grid(True, alpha=0.3)
    
    # Добавляем среднее количество символов на слово в виде текста
    avg_chars_per_word = np.mean([chars/words for chars, words in zip(prompt_chars, prompt_words)])
    plt.text(0.02, 0.98, 
            f'Среднее количество символов на слово: {avg_chars_per_word:.1f}',
            transform=plt.gca().transAxes,
            verticalalignment='top',
            fontsize=28,
            bbox=dict(boxstyle='round', facecolor='white', alpha=0.8))
    
    plt.tight_layout()
    plt.savefig(os.path.join(output_dir, 'prompt_length_histogram.png'), dpi=200)
    plt.close()

    # График количества коммитов по времени
    df_commits = pd.DataFrame(commits)
    if not df_commits.empty:
        # Фильтруем коммиты до 2025-01-10
        df_commits = df_commits[df_commits['datetime'] < pd.Timestamp('2025-01-10')]
        # Группируем по дате и часу
        df_commits['datetime_hour'] = df_commits['datetime'].dt.floor('H')
        commits_per_hour = df_commits.groupby('datetime_hour').size()
        
        # Считаем количество измененных строк по часам (только PHP и YAML)
        df_commits['changes'] = df_commits['php_yaml_additions'] + df_commits['php_yaml_deletions']
        changes_per_hour = df_commits.groupby('datetime_hour')['changes'].sum()
        
        # Создаем график с двумя осями
        fig, ax1 = plt.subplots(figsize=(30,16))
        
        # Основная ось для количества коммитов
        plt.grid(True, axis='y')
        ax1.bar(range(len(commits_per_hour)), commits_per_hour.values,
               color='mediumseagreen', edgecolor='black', width=0.8, alpha=0.7)
        ax1.set_title('Количество коммитов и измененных строк по времени\n(учитываются только PHP и YAML файлы)', 
                     fontsize=36, pad=20)
        ax1.set_xlabel('Дата и время', fontsize=28, labelpad=20)
        ax1.set_ylabel('Количество коммитов', color='mediumseagreen', fontsize=28)
        ax1.tick_params(axis='y', labelcolor='mediumseagreen', labelsize=24)
        
        # Устанавливаем метки времени для каждого часа (показываем каждую 12-ю метку)
        xticks_pos = range(len(commits_per_hour))[::3]
        xticks_labels = [t.strftime('%Y-%m-%d\n%H:00') for t in commits_per_hour.index][::3]
        plt.xticks(xticks_pos, xticks_labels, rotation=45, ha='center', fontsize=24)
        
        # Дополнительная ось для количества строк
        ax2 = ax1.twinx()
        ax2.scatter(range(len(changes_per_hour)), changes_per_hour.values, 
                   color='red', s=100, alpha=0.7, marker='o')
        # Добавляем вертикальные линии от точек к оси X для лучшей читаемости
        for x, y in zip(range(len(changes_per_hour)), changes_per_hour.values):
            ax2.vlines(x, 0, y, colors='red', alpha=0.2)
        ax2.set_ylabel('Количество измененных строк (PHP+YAML)', color='red', fontsize=28)
        ax2.tick_params(axis='y', labelcolor='red', labelsize=24)
        
        # Настраиваем отступы
        plt.subplots_adjust(bottom=0.2, top=0.9, right=0.9, left=0.1)
        
        plt.savefig(os.path.join(output_dir, 'commits_per_day.png'), dpi=200, bbox_inches='tight', pad_inches=0.3)
        plt.close()
 

        # Новый график - анализ строк кода
        # Группируем добавления и удаления по часам
        additions_per_hour = df_commits.groupby('datetime_hour')['php_yaml_additions'].sum()
        deletions_per_hour = df_commits.groupby('datetime_hour')['php_yaml_deletions'].sum()
        
        # Отдельно для тестов и основного кода
        test_additions_per_hour = df_commits.groupby('datetime_hour')['test_additions'].sum()
        test_deletions_per_hour = df_commits.groupby('datetime_hour')['test_deletions'].sum()
        main_additions_per_hour = df_commits.groupby('datetime_hour')['main_additions'].sum()
        main_deletions_per_hour = df_commits.groupby('datetime_hour')['main_deletions'].sum()
        
        # Рассчитываем общее количество строк (кумулятивная сумма)
        test_lines = (test_additions_per_hour - test_deletions_per_hour).cumsum()
        main_lines = (main_additions_per_hour - main_deletions_per_hour).cumsum()
        
        # Создаем график с двумя осями Y
        fig, ax1 = plt.subplots(figsize=(30,16))
        
        # Основная ось для добавлений и удалений
        ax1.bar(range(len(additions_per_hour)), additions_per_hour.values,
               color='lightgreen', edgecolor='darkgreen', alpha=0.7, label='Добавлено строк')
        ax1.bar(range(len(deletions_per_hour)), -deletions_per_hour.values,
               color='salmon', edgecolor='darkred', alpha=0.7, label='Удалено строк')
        
        ax1.set_title('Анализ изменений кода по времени\n(учитываются только PHP и YAML файлы)', 
                     fontsize=36)
        ax1.set_xlabel('Дата и время', fontsize=28)
        ax1.set_ylabel('Количество измененных строк', fontsize=28)
        ax1.grid(True, alpha=0.3)
        ax1.legend(loc='upper left', fontsize=24)
        ax1.tick_params(labelsize=24)
        
        # Дополнительная ось для общего количества строк
        ax2 = ax1.twinx()
        ax2.plot(range(len(main_lines)), main_lines.values, 
                color='blue', linewidth=3, label='Строк в основном коде')
        ax2.plot(range(len(test_lines)), test_lines.values, 
                color='purple', linewidth=3, label='Строк в тестах')
        
        ax2.set_ylabel('Общее количество строк', color='blue', fontsize=28)
        ax2.tick_params(axis='y', labelcolor='blue', labelsize=24)
        ax2.legend(loc='upper right', fontsize=24)
        
        # Устанавливаем метки времени для каждого часа (показываем каждую 8-ю метку)
        xticks_pos = range(len(additions_per_hour))[::8]
        xticks_labels = [t.strftime('%Y-%m-%d\n%H:00') for t in additions_per_hour.index][::8]
        plt.xticks(xticks_pos, xticks_labels, rotation=0, ha='center', fontsize=24)
        
        # Увеличиваем отступ снизу для подписей
        plt.subplots_adjust(bottom=0.2)
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, 'code_changes.png'), dpi=200)
        plt.close()

    # Если промпты классифицированы, строим распределение типов промптов
    if prompts and any('type' in p for p in prompts):
        prompt_types = [p.get('type', 'not_classified') for p in prompts]
        df_prompts = pd.DataFrame(prompt_types, columns=['type'])
        prompt_type_counts = df_prompts['type'].value_counts()
        
        # Рассчитываем проценты
        total = prompt_type_counts.sum()
        percentages = (prompt_type_counts / total * 100).round(1)
        
        plt.figure(figsize=(24,12))
        ax = prompt_type_counts.plot(kind='bar', color='#7CB9E8')
        plt.title('Распределение типов промптов', fontsize=36)
        plt.xlabel('Тип промпта', fontsize=28)
        plt.ylabel('Количество промптов', fontsize=28)
        plt.xticks(rotation=45, ha='right', fontsize=24)
        plt.yticks(fontsize=24)
        
        # Добавляем подписи только с процентами
        for i, v in enumerate(prompt_type_counts):
            percentage = percentages[i]
            ax.text(i, v, f'{percentage}%', 
                   ha='center', va='bottom', fontsize=24)
        
        plt.grid(True, axis='y', alpha=0.3)
        plt.tight_layout()
        plt.savefig(os.path.join(output_dir, 'prompt_types.png'), dpi=200)
        plt.close()

    print(f"Graphs saved to {output_dir}/")


def generate_analysis_readme(prompts, commits, output_dir='output'):
    """Генерирует README.md с результатами анализа."""
    total_prompts = len(prompts)
    total_commits = len(commits)
    
    # Подсчет статистики по промптам
    prompt_types = {}
    prompt_words = []
    prompt_chars = []
    for p in prompts:
        ptype = p.get('type', 'not_classified')
        prompt_types[ptype] = prompt_types.get(ptype, 0) + 1
        prompt_words.append(p['word_count'])
        prompt_chars.append(p['char_count'])
    
    # Подсчет статистики по коммитам
    if commits:
        df_commits = pd.DataFrame(commits)
        df_commits = df_commits[df_commits['datetime'] < pd.Timestamp('2025-01-10')]
        commits_by_date = df_commits.groupby(df_commits['datetime'].dt.date).size()
        commits_by_hour = df_commits.groupby(df_commits['datetime'].dt.hour).size()
        avg_commits_per_day = commits_by_date.mean()
        most_active_hour = commits_by_hour.idxmax()
        total_additions = df_commits['additions'].sum()
        total_deletions = df_commits['deletions'].sum()
        avg_changes_per_commit = (total_additions + total_deletions) / total_commits if total_commits > 0 else 0
        
        # Подсчет потраченных часов (если есть коммит в пределах часа - считаем этот час)
        unique_hours = df_commits['datetime'].dt.floor('H').nunique()

        # Подсчет итогового количества строк в src и tests
        total_main_additions = df_commits['main_additions'].sum()
        total_main_deletions = df_commits['main_deletions'].sum()
        total_test_additions = df_commits['test_additions'].sum()
        total_test_deletions = df_commits['test_deletions'].sum()
        
        total_main_lines = total_main_additions - total_main_deletions
        total_test_lines = total_test_additions - total_test_deletions
    
    # Формируем markdown
    readme_content = f"""# Анализ разработки с помощью Cursor

## Общая статистика

- Всего промптов: {total_prompts}
- Всего коммитов: {total_commits}
- Среднее количество промптов на коммит: {total_prompts / total_commits:.1f}
- Потрачено часов на проект: {unique_hours} (учитываются часы, в которых были коммиты)
- Среднее количество слов в промпте: {sum(prompt_words) / total_prompts:.1f}
- Среднее количество символов в промпте: {sum(prompt_chars) / total_prompts:.1f}
- Среднее количество коммитов в день: {avg_commits_per_day:.1f}
- Самый активный час для коммитов: {most_active_hour}:00
- Среднее количество изменений на коммит: {avg_changes_per_commit:.1f} строк
- Количество строк в основном коде (src): {total_main_lines:,} строк
- Количество строк в тестах (tests): {total_test_lines:,} строк
- Соотношение тестов к коду: {(total_test_lines / total_main_lines * 100):.1f}%

## Распределение типов промптов

"""
    
    # Добавляем статистику по типам промптов
    total_classified = sum(prompt_types.values())
    readme_content += "| Тип | Количество | Процент |\n"
    readme_content += "|-----|------------|----------|\n"
    for ptype, count in sorted(prompt_types.items(), key=lambda x: x[1], reverse=True):
        percentage = (count / total_classified) * 100
        readme_content += f"| {ptype} | {count} | {percentage:.1f}% |\n"

    # Добавляем графики
    readme_content += """
## Графики

### Распределение длины промптов
![Распределение длины промптов](output/prompt_length_histogram.png)

### Распределение типов промптов
![Распределение типов промптов](output/prompt_types.png)

### Активность коммитов по времени
![Активность коммитов по времени](output/commits_per_day.png)

### Анализ изменений кода
![Анализ изменений кода](output/code_changes.png)

## Выводы
"""

    # Формируем выводы отдельно для корректной подстановки значений
    most_frequent_type = max(prompt_types.items(), key=lambda x: x[1])[0]
    most_frequent_count = prompt_types[most_frequent_type]
    avg_words = sum(prompt_words) / total_prompts

    readme_content += f"""
1. **Промпты**:
   - Наиболее частый тип промптов: {most_frequent_type} ({most_frequent_count} промптов)
   - Средняя длина промпта составляет {avg_words:.1f} слов

2. **Коммиты**:
   - В среднем {avg_commits_per_day:.1f} коммитов в день
   - Наибольшая активность наблюдается в {most_active_hour}:00
   - Среднее количество изменений в коммите: {avg_changes_per_commit:.1f} строк
"""

    # Сохраняем README
    with open('ANALYSIS.md', 'w', encoding='utf-8') as f:
        f.write(readme_content)
    
    print("Анализ сохранен в ANALYSIS.md")


def main():
    parser = argparse.ArgumentParser(description='Анализ данных разработки с помощью Cursor')
    parser.add_argument('--cursor-log', type=str, default='cursor-log.md', help='Путь к файлу cursor-log.md')
    parser.add_argument('--interactive-prompts', action='store_true', help='Включить интерактивную классификацию промптов с типом "other"')
    args = parser.parse_args()

    print("Начинаю анализ cursor-log...")
    prompts = analyze_cursor_log(args.cursor_log)
    
    if args.interactive_prompts and prompts:
        prompts = classify_prompts(prompts)

    print("Анализ истории git...")
    commits = analyze_git_log()

    print("Построение графиков...")
    plot_graphs(prompts, commits)

    print("Генерация отчета...")
    generate_analysis_readme(prompts, commits)

    print("Анализ завершён.")


if __name__ == '__main__':
    main() 