# neo
Client + server for exploits distribution used for Attack-Defence CTF competitions.

Definitely not a botnet. 

Neo это 

### Запуск клиента

1. Скачайте архив с клиентом из релизов под вашу платформу.
2. Измените конфиг: поставьте IP и порт сервера, ключ.  
По-умолчанию клиент будет искать файл `config.yml` в текущей директории, можно кастомизировать это с
помощью флага `--config`).
3. Запустите `./neo run`. Вы можете указать флаг `-j` (`--jobs`), чтобы управлять максимальным одновременным количеством запускаемых эксплойтов (по-умолчанию равно количеству ядер).

#### Как добавить эксплойт ?

Запустите `./neo add <путь к эксплойту>`. 

Клиент проведет валидацию и отправит эксплойт на сервер. 

Вы можете кастомизировать название (ID) эксплойта с помощью флага `--id`.

Neo также поддерживает опцию `--dir`, которая позволяет отправить в качестве эксплойта целую папку.
В таком случае вам нужно запустить `./neo add --id new_sploit --dir path_to_folder/sploit.py`.
Все содержимое папки path_to_folder будет отправлено другим клиентам, эксплойт "sploit.py"
будет запущен внутри данной директории.

### Запуск сервера

1. Скачайте архив с клиентом из релизов под вашу платформу.
2. Измените конфиг:  
По-умолчанию клиент будет искать файл `config.yml` в текущей директории, можно кастомизировать это с
помощью флага `--config`). Вам нужно указать публичный адрес фермы для отправки флагов, ключ, IP команд. 
**Вам не нужно перезапускать сервер при изменении конфига**.
3. Запустите `./neo_server`.

