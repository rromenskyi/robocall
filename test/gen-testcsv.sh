#!/bin/bash

#!/bin/bash

# Проверяем, был ли предоставлен аргумент
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <number_of_lines>"
    exit 1
fi

number_of_lines=$1

# Создаем или очищаем файл output.txt перед записью
> output.csv

for (( i=1; i<=$number_of_lines; i++ ))
do
    # Генерируем случайное число от 1 до 10000
    random_number=$((RANDOM % 10000 + 1))

    # Генерируем номер телефона в формате 380XXXXXXXXX
    phone_number="380"
    for j in {1..9}
    do
      phone_number+=$((RANDOM % 10))
    done

    # Записываем строку в файл
    echo "${random_number};${phone_number};test;" >> output.csv
done

echo "$number_of_lines строк записано в output.csv!"
