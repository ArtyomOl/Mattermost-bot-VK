# Mattermost-bot-VK

Бот для создания голосований в mattermost и управлении ими. 
Создан в качестве ответа на тестовое задание на стажировку в компанию VK.  
Автор - Оленев Артём  

## Установка:
Для уствновки скланируйте данный репозиторий:  
`git clone https://github.com/ArtyomOl/Mattermost-bot-VK `  

## Запуск:
- Замените в файле config.go константы MattermostToken, channelID и userID на свои соответствующие значения  
- Запустите tarantool и mattermost  
- Соберите контейнер с ботом командой
`docker compose build mattermost-bot`  
- Запустите бота командой `docker compose up -d mattermost-bot`  

#### Либо (без использования Docker):
- Замените в файле config.go константы MattermostToken, channelID и userID на свои соответствующие значения
- Запустите tarantool и mattermost
- Выполните команду `go run main.go`

## Использование:
Используйте в своем канале mattermost следующие команды:  
1. Создание голосования:  
`/vote create "<Question>" "<Option 1>" "<Option 2>" ... "<Option n>"`  
После создания голосования в чат выведется ссобщение с id голосования, которое будет необходимо для последующей работы с ним.  
2. Голосование:  
`/poll vote <poll_id> <option>`  
3. получение результатов:  
`/poll results <poll_id>`
4. Завершение голосования:  
`/poll end <poll_id>`  
5. Удаление голосования:  
`/poll delete <poll_id>`