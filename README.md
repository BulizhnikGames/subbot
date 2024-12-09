[О боте](#о-боте)

[Как это работает](#как-это-работает)

[Ограничения](#ограничения)

# О [боте](https://t.me/subfwdbot)

## Как использовать

Добавьте бота в группу, которую хотите подписать на телеграм канал(ы) после чего используйте команду /sub (если в группе есть несколько тем, то эту команду стоит использовать в той теме, куда хотелось бы получать пересланные сообщения)

## Что подразумевается под подпиской группы на канал

Бот будет пересылать (если это возможно) новые посты из телеграм канала во все группы, которые на него подписаны. 
Так же бот будет пересылать посты из канала, при их редактировании (за исключением некоторых случаев).

# Как это работает

Система на самом деле чуть сложнее, чем бот, просто подписывающийся на телеграм каналы, потому что боты не могут подписываться на них.

## Обход запрета на подписывания ботов

Для этого отдельно от бота создаётся [фетчер](https://github.com/BulizhnikGames/subbot/tree/master/fetcher), который подписывается на нужные каналы, пересылает посты из них боту, а бот уже пересылает их группам-подписчикам.

### Более подробно про фетчер

Фетчер представляется собой обычный аккаунт телеграм, управление которым отдаётся программе, а не человеку.

При увеличении количества пользователей бота, скорее всего обойтись одним фетчером не получиться (у телеграм аккаунтов ограничение в 500 подписок на каналы). Поэтому система позволяет очень легко добавлять новые фетчеры, для этого надо просто создать новый телеграм аккаунт и запустить программу фетчера, введя в ней данные от нового аккаунта.

Фетчер вполне себе можно использовать как отдельный проект, так как он позволяет сохранять посты из телеграм каналов в их оригинальном виде (если отредактировать оригинальное сообщение, то его пересланная версия не измменится).

### Как бот отличает фетчеров от обычных пользователей?

Фетчеры перенаправляют посты из каналов боту с помощью пересылки их в личные сообщения боту, поэтому бот должен уметь отличать обычного человека, которые решил переслать ему пост в лс от фетчера.

Для этого у бота [база данных](https://github.com/BulizhnikGames/subbot/blob/master/bot/db/migrations/002_fetchers.sql), в которую фетчеры добавляются при их запуске.

### Как бот коммуницирует с фетчерами?

Бот коммуницирует с фетчерами с помощью протокола http

Коммуникация работает в обе стороны:
1. Фетчеры отправляются боту запрос на регистрацию в базе данных фетчеров при запуске
2. Бот отправляет фетчерам запросы на подписку/отписку от канала.

# Ограничения

1. Бот не умеет подписываться за приватные каналы
2. Бот пересылает сообщения из каналов с запретом на пересылку в таком формате:
> Новое сообщение в канале @*Имя канала*:
> https://t.me/channel/id (ссылка на сообщение)
3. Бот игнорирует редактирование постов, на которых есть кнопки, так как скорее всего это кнопка для участия в розыгрыше, которая скорее всего отображает количество участников, а значит при каждом новом участнике в розыгрыше это будет новый апдейт редактирования, что очень много.
4. Бот считает за редактирование только редактирование текста поста (изменение фото или что-то другое игнорируется).
