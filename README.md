# Project_template

# Задание 1. Анализ и планирование

### 1. Описание функциональности монолитного приложения

По условию система обеспечивает включение/выключение отопления и просмотр температуры. Предоставленный код реализует учёт датчиков и получение температурных показаний. Изменение значения и статуса датчика затрагивает только запись в PostgreSQL; отправка управляющих команд оборудованию отсутствует.

**Границы анализа и принятые допущения**

Описание системы в задании и предоставленная реализация не полностью соответствуют друг другу. В условии заявлено управление отоплением, однако код содержит операции чтения показаний и изменения записей о датчиках в PostgreSQL без отправки команд оборудованию. Поэтому ниже различаются факты, подтверждённые кодом, и допущения для архитектурного анализа.

- **Подтверждено кодом:** монолит запрашивает показания у внешнего temperature API. Операции изменения значения и статуса датчика обновляют только БД.
- **Управление температурой не реализовано:** введённые через API значения сохраняются в PostgreSQL и читаются обработчиками GET, но не используются для регулирования отопления. При обновлении записи команды или целевые значения во внешнюю систему не передаются. Фоновой логики, которая читает эти значения и управляет оборудованием, в коде нет.
- **Ручное изменение данных вместо управления оборудованием:** API позволяет вручную менять значение и статус записи о датчике. . Модель не разделяет измеренную и целевую температуру; при успешном запросе к внешнему temperature API сохранённые значения подменяются его показаниями в ответе GET. 
- **Подтверждено кодом:** аутентификация, авторизация и разграничение доступа по ролям отсутствуют. Все зарегистрированные маршруты доступны любому клиенту, имеющему сетевой доступ к HTTP-серверу. Фактическая доступность из интернета зависит от конфигурации развёртывания и по коду не устанавливается.
- **Допущение о назначении API:** конечному пользователю предназначены только явно описанные в задании сценарии — включение/выключение отопления и просмотр температуры. Остальные операции считаю административными или интеграционными: учёт и изменение записей о датчиках — для администраторов, чтение и обновление данных в рамках возможных интеграций — для внешних сервисов. Наличие таких интеграций кодом не подтверждено. Точное соответствие пользовательских сценариев конкретным endpoint в условии не задано; изменение записи в БД само по себе не считается подтверждённой командой оборудованию.

Такое распределение доступа принято для анализа и проектирования; в текущем коде оно не обеспечивается. Фактические права вызывающих сторон и наличие внешней управляющей автоматики неизвестны.

**Управление отоплением:**

- По условию пользователь должен иметь возможность включать и выключать отопление.
- В предоставленном коде отправка команд оборудованию отсутствует: изменение `value` и `status` обновляет только запись в БД.
- Регулирование температуры и автоматическое поддержание заданного значения не реализованы. Доступно ручное изменение данных о датчике, но не подтверждённое управление физическим устройством.

**Мониторинг температуры:**

- API предоставляет список датчиков и сведения об отдельном датчике из БД.
- Для температурных датчиков при чтении выполняется синхронный запрос показаний во внешний API; при успехе они подставляются в ответ без сохранения в БД, при ошибке возвращаются данные из БД.
- Температуру по локации можно запросить напрямую у внешнего API через монолит. Фонового опроса и автоматических обновлений для клиента в коде нет.

### 2. Анализ архитектуры монолитного приложения

Перечислите здесь основные особенности текущего приложения: какой язык программирования используется, какая база данных, как организовано взаимодействие между компонентами и так далее.
Язык Go, БД Postgresql, есть внешнее api  - через которое запрашиваются данные по температуре. Внешнее api, по-видимому, сейчас само пулит данные по датчику (в задании написано управление от сервера к датчику, мутирующих запросов на apps/smart_home/services/temperature_service.go я не нашел).
Как создается датчик для меня загадка, так как в `CreateSensor` нет передачи идентификатора, в базе location не unique, вероятно внешняя система может мапить их по локации (но в базе локация не unique)

### 3. Определение доменов и границы контекстов

Предметная область системы по условию — управление умным домом. В текущей реализации представлены две функциональные области: учёт датчиков и получение температурных показаний.

#### 1. Учёт датчиков

- Добавление записи о температурном датчике.
- Изменение данных датчика.
- Удаление записи о датчике.
- Просмотр списка датчиков и сведений об отдельном датчике.

Эта область отвечает за сведения о датчиках, хранящиеся в БД монолита. Добавление или удаление записи не означает регистрацию или отключение физического устройства во внешней системе.

#### 2. Мониторинг температуры в помещениях

- Получение текущей температуры через внешний API по идентификатору датчика или локации.
- Подстановка полученных показаний в ответ при просмотре температурных датчиков.

### **4. Проблемы монолитного решения**
- В GetTemperatureByLocation почему-то нет похода в базу, все данные, включая локацию и описание, берутся из внешнего API. Смешаны домены учёта датчиков и получения показаний, не ясно кто отвечает за данные датчика.
- Заявленное в условии управление отоплением не реализовано: обновление значения и статуса меняет только БД, не отправляет команды оборудованию и не запускает регулирование температуры.
- Отсутвует слой бизнес логики, все смешано в хендлерах 
- Завязка на конкретные реализации PostgreSQL, API датчиков (Backing Services)
- Текущий endpoint GetTemperatureByLocation достанет только один датчик
- Не ясно как создаются дачтики, идентификатор создается в базе, при этом сами датчики не создаются во внешней api и он не заполняется (судя по контракту apps/smart_home/models/SensorCreate), как тогда работает GetTemperatureByID, либо оно работает так же по polling и датчик себе записывает идентификатор по локации (тогда проблема  тем что на локации реально один датчик)
- Получение актуальной температуры синхронно зависит от доступности внешнего API. При чтении списка или отдельного датчика ошибка внешнего API приводит к возврату сохранённых в PostgreSQL, потенциально устаревших данных; при запросе температуры по локации — к HTTP 500. Успешно полученные показания подставляются в ответ, но не сохраняются в БД. Чтобы отделить обработку пользовательских запросов от доступности внешнего API, можно вынести опрос в отдельный процесс, сохранять показания и отдавать их из БД с указанием времени измерения и актуальности. Это позволит читать последние сохранённые показания при сбоях внешнего API, но не обеспечит получение свежих данных во время его недоступности.
- При получении списка происходит последовательный fetch запросов (с таймаутом в 10 секун) по устройствам при 100 приборах это может при недосутпаности внешнего api привести к ожиданию порядка 1000 секунд, при чем которые никак не обрабатывают отмену клиентского запроса (будет просто висеть в background на polling).
- Запросы к базе идут без внешнего context handler'а, а заменяются background, что так же не обрывает запрос к базе в случае отмены на клиенте.
- Отсутствует модель управления доступами пользователей, ни авторизации, ни пользовательской таблицы в нем нет. В предоставленном коде все операции работают с общей таблицей датчиков, по всей видимости, пользователи могут получать и управлять чужими данными.

### 5. Визуализация контекста системы — диаграмма С4

```puml
@startuml
title WarmHouse Context Diagram

top to bottom direction

!include <C4/C4_Context>

Person(user, "User", "A user of the Warm House system")
Person(admin, "Administrator", "An administrator managing the system")
System(WarmHouseSystem, "Warm House System", "Enables remote heating control and temperature monitoring in homes")

System_Ext(temperatureApi, "Temperature API", "Provides temperature readings by sensor ID or location")

Rel(user, WarmHouseSystem, "Uses the system","HTTP/JSON")
Rel(admin, WarmHouseSystem, "Manages the system","HTTP/JSON")
Rel(WarmHouseSystem,temperatureApi, "Fetches temperature readings", "HTTP/JSON")

@enduml
```
# Задание 2. Проектирование микросервисной архитектуры

В этом задании вам нужно предоставить только диаграммы в модели C4. Мы не просим вас отдельно описывать получившиеся микросервисы и то, как вы определили взаимодействия между компонентами To-Be системы. Если вы правильно подготовите диаграммы C4, они и так это покажут.

**Диаграмма контейнеров (Containers)**

```puml
@startuml
title WarmHouse - Container Diagram

!include <C4/C4_Container>

top to bottom direction

Person(user, "User", "Uses the WarmHouse system")
System_Boundary(WarmHouseSystem, "WarmHouse system") {
    Container(WebUI, "Web UI", "JavaScript", "Smart home interface")
    Container(AuthService, "Auth Service", "Go", "Users, authentication, sessions and tokens")
    Container(BFF, "Backend For Frontend", "Go", "Owns access mappings and user operations; serves provisional pending/validating/active/failed UI cards")
    Container(AutomationService, "Automation Service", "Go", "Scenario run state machines and schedules referencing internal device IDs")

    ContainerQueue(automationQueue, "Automation Internal Events", "YDB Topics", "Private run requests; durable scenario scheduling")

    Boundary(deliveryPool, "Delivery Service", "Service group") {
        Container(DeviceManagement, "Device Management", "Go", "Owns device IDs, type references and lifecycle; activates devices only after successful Delivery verification")
        Container(Monitoring, "Monitoring", "Go", "Owns type monitoring profiles, polling schedules, metrics and availability; starts monitoring on DeviceCreated")
        Boundary(deliveryRuntime, "Command Delivery runtime", "Application and dedicated infrastructure") {
            Container(Delivery, "Command Delivery", "Go", "Verifies API connections; executes durable tasks with fenced leases and retries; owns provider IDs and soft-delete checks; synchronous video sessions")
            ContainerQueue(deliveryQueue, "Delivery Task Queue", "YDB Topics", "Private task IDs; task + queue message commit before CreateTask returns")
            ContainerDb(deliveryLimiter, "Delivery Rate Limiter", "Redis", "Dedicated shared RPS quota state across Delivery replicas, scoped by integration/account")
        }
    }

    ContainerQueue(eventBus, "Domain Events", "YDB Topics", "Lifecycle, telemetry, scenario and task results; original operation/command IDs for correlation")
    ContainerDb(database, "Smart Home Database", "YDB Cluster", "Separate service-owned tables; atomic table/topic transactions within the same YDB database")
}
System_Ext(partnerApis, "Provider APIs", "Device verification, control, telemetry and video sessions")

Rel(user, WebUI, "Uses")
Rel(WebUI, BFF, "Connects provider accounts; selects devices; manages scenarios and commands", "HTTPS / JSON")
Rel(WebUI, AuthService, "Signs in and manages sessions", "HTTPS / JSON")
Rel_L(BFF, AuthService, "Validates token; receives user_id", "gRPC / Protobuf")
Rel(BFF, AutomationService, "Manages authorized scenarios", "gRPC / Protobuf")
Rel(BFF, DeviceManagement, "Creates/deletes devices; tracks pending status", "gRPC / Protobuf")
Rel(BFF, Delivery, "CreateTask(CreateConnection/commands); lists opaque candidates; creates video sessions", "gRPC / Protobuf")
Rel(WebUI, partnerApis, "Receives authorized camera stream", "WebRTC")

Rel_R(Monitoring, Delivery, "Bounded batch submission via CreateTask(ReadState); same command_id on retry", "gRPC / Protobuf")
Rel(AutomationService, Delivery, "CreateTask for scenario actions; retries durable intents", "gRPC / Protobuf")
Rel_D(AutomationService, automationQueue, "Atomically schedules runs; publishes/consumes run requests", "YDB SDK / Protobuf")
Rel_D(Delivery, deliveryQueue, "Enqueues/consumes task IDs; acknowledges durable task handling", "YDB SDK / Protobuf")
Rel_R(DeviceManagement, Delivery, "Creates verification/deactivation task; receives task_id after durable acceptance", "gRPC / Protobuf")
Rel_R(Delivery, partnerApis, "Verifies devices, executes commands, reads state and creates video sessions", "HTTPS / JSON")
Rel(Delivery, deliveryLimiter, "Provider Adapters atomically acquire shared RPS quota", "Redis protocol")

Rel(BFF, eventBus, "Consumes lifecycle, state, command results and scenario events", "YDB SDK / Protobuf")
Rel(AutomationService, eventBus, "Publishes scenario/run events; consumes state, DeviceDeleted and task results", "YDB SDK / Protobuf")
Rel(DeviceManagement, eventBus, "Consumes verification/deactivation results; publishes lifecycle success and correlated operation failure/unknown", "YDB SDK / Protobuf")
Rel(Monitoring, eventBus, "Consumes lifecycle/read results; publishes telemetry events", "YDB SDK / Protobuf")
Rel(Delivery, eventBus, "Publishes task results with original operation/command IDs", "YDB SDK / Protobuf")

Rel(AuthService, database, "Users and sessions", "YDB SDK")
Rel(BFF, database, "Operation owners, leased durable intents, provisional UI cards, final ownership mappings and projections", "YDB SDK")
Rel(AutomationService, database, "Scenarios, schedules, run event journal, state projections and command correlations", "YDB SDK")
Rel(DeviceManagement, database, "Device IDs, type references, lifecycle and pending operations", "YDB SDK")
Rel(Monitoring, database, "Type profiles, schedules, leased pending polls and correlations, metrics and availability", "YDB SDK")
Rel(Delivery, database, "Provider type/command catalog, verified connections, candidate mappings, targets, task states and results", "YDB SDK")


@enduml
```

**Диаграмма компонентов (Components)**

### Auth Service

```puml
@startuml
title WarmHouse - Auth Service Components

top to bottom direction

!include <C4/C4_Component>

Container(WebUI, "Web UI", "JavaScript", "Registration, sign-in and session management")
Container(BFF, "Backend For Frontend", "Go", "Owns user-to-object mappings and authorizes user requests using the validated user ID")

Container_Boundary(AuthService, "Auth Service") {
    Component(httpApi, "Public HTTP API", "Go / HTTP", "Validates requests for registration, sign-in, refresh and sign-out")
    Component(grpcApi, "Internal gRPC API", "Go / gRPC", "Authenticates calling service; delegates token validation to Session Controller; returns user identity")
    Component(accounts, "Account Controller", "Go", "Creates user IDs, registers users, hashes passwords, verifies credentials and initiates sessions")
    Component(sessions, "Session Controller", "Go", "Issues opaque access and refresh tokens, validates expiry and revocation, rotates refresh tokens and revokes sessions")
    Component(usersRepo, "User Repository", "Go / YDB SDK", "Stores user IDs, unique login identifiers and password hashes")
    Component(sessionsRepo, "Session Repository", "Go / YDB SDK", "Stores session owners, token hashes, expiration and revocation state; atomically rotates refresh tokens")
}

ContainerDb(database, "Smart Home Database", "YDB", "Auth-owned users and sessions tables")

Rel(WebUI, httpApi, "Registers, signs in, refreshes tokens and signs out", "HTTPS / JSON")
Rel_D(BFF, grpcApi, "Validates access token; receives user ID or rejection", "gRPC / Protobuf")
Rel(grpcApi, sessions, "Validates access tokens against active sessions", "Go calls")
Rel(httpApi, accounts, "Registers users and verifies sign-in credentials", "Go calls")
Rel(httpApi, sessions, "Refreshes tokens and revokes sessions", "Go calls")

Rel_L(accounts, sessions, "Creates a session after successful sign-in", "Go calls")
Rel(accounts, usersRepo, "Creates users and reads credential records", "Go calls")
Rel(sessions, sessionsRepo, "Creates, reads, rotates and revokes sessions", "Go calls")
Rel(usersRepo, database, "Reads and writes users; enforces login uniqueness", "YDB SDK")
Rel(sessionsRepo, database, "Reads and writes sessions transactionally", "YDB SDK")
@enduml
```

### Device Management

```puml
@startuml
title WarmHouse - Device Management Components
!include <C4/C4_Component>
top to bottom direction

Container(BFF, "BFF", "Go", "Checks user ownership; requests device lifecycle operations")
ContainerQueue(results, "Delivery Task Results", "YDB Topic", "Verification/deactivation outcomes with device_id, operation_id and task_id")

Container_Boundary(DeviceManagement, "Device Management") {
    Component(grpcApi, "Internal gRPC API", "Go / gRPC", "Authenticates callers; BFF create/delete/status; consistent device snapshot for Automation bootstrap")
    Component(resultConsumer, "Delivery Result Consumer", "Go / YDB SDK", "Consumes verification/deactivation results; passes transaction and offsets to lifecycle processing")
    Component(lifecycle, "Device Lifecycle Controller", "Go", "Owns canonical types/commands; publishes new device snapshots on effective capability changes; allocates internal IDs; stores pending operations with connection/candidate IDs; soft-deletes after Delivery confirmation")
    Component(recovery, "Pending Operation Dispatcher", "Go", "Claims persisted unsubmitted operations after restart; retries with stable operation_id until accepted")
    Component(deliveryClient, "Delivery Client", "Go / gRPC", "Sends verification/deactivation RPCs with supplied operation_id; returns acceptance/errors; retry policy belongs to dispatcher")
    Component(deviceRepo, "Device Repository", "Go / YDB SDK", "Devices: internal ID, name, verified type, effective canonical commands/parameter constraints and supported metrics, lifecycle/version and soft delete. Operations: operation/task IDs, status and retry payload; connection/candidate IDs only in create payload")
    Component(eventPublisher, "Lifecycle Event Publisher", "Go / YDB SDK", "Publishes versioned full DeviceCreated/DeviceUpdated capability snapshots or DeviceDeleted tombstones with state; failed/unknown operation outcomes remain separate")
}

Container(Delivery, "Command Delivery", "Go", "Commits task + private queue message before returning task_id; verifies/deactivates targets asynchronously")
ContainerDb(database, "Device Management Tables", "YDB", "Devices, canonical type/command definitions, effective capabilities and pending operations; same database as lifecycle topics")
ContainerQueue(events, "Device Lifecycle Events", "YDB Topic", "BFF consumes success and failed/unknown operation outcomes; Automation projects DeviceCreated/DeviceUpdated/DeviceDeleted; Monitoring consumes create/delete to update schedules")

Rel_D(BFF, grpcApi, "Creates/deletes devices; reads operation status", "gRPC / Protobuf")
Rel_U(resultConsumer, results, "Consumes results with a shared consumer", "YDB SDK / Protobuf")
Rel_D(grpcApi, lifecycle, "Validates and dispatches lifecycle requests", "Go calls")
Rel_D(resultConsumer, lifecycle, "Applies result and commits offset in one transaction", "Go calls / transaction")
Rel_D(recovery, deviceRepo, "Claims pending submissions; conditionally stores acceptance", "Go calls / transaction")
Rel_D(recovery, lifecycle, "Applies terminal submission rejection through lifecycle handler", "Go calls")
Rel_D(recovery, deliveryClient, "Resubmits outside DB transactions", "Go calls")
Rel_D(lifecycle, deliveryClient, "Submits task after persisting pending operation", "Go calls; outside DB transaction")
Rel_D(lifecycle, deviceRepo, "Persists pending state; applies matching results and verified type/capabilities; updates affected device snapshots when type definitions change", "Go calls / transaction")
Rel_D(lifecycle, eventPublisher, "Atomically publishes successful lifecycle or failed/unknown operation outcome with state", "Go calls / transaction")
Rel_D(deviceRepo, database, "Reads/writes device lifecycle and operations", "YDB SDK")
Rel_D(eventPublisher, events, "Publishes correlated lifecycle success or operation failure/unknown; no false DeviceCreated/DeviceDeleted", "YDB SDK / transaction")
Rel_D(deliveryClient, Delivery, "CreateTask: verify/deactivate target; stable operation_id and internal device_id", "gRPC / Protobuf")

@enduml
```

### Command Delivery

```puml
@startuml
title WarmHouse - Command Delivery Components
!include <C4/C4_Component>
top to bottom direction
skinparam nodesep 60
skinparam ranksep 65
skinparam linetype spline

Container(Management, "Device Management", "Go", "Requests verification/deactivation tasks with device_id and operation_id")
Container(BFF, "BFF", "Go", "Checks ownership; persists command intents; requests tasks or video sessions")
Container(Automation, "Automation Service", "Go", "Persists action intents and retries CreateTask with stable command IDs")
Container(Monitoring, "Monitoring", "Go", "Persists polling intents and retries CreateTask(ReadState)")

Container_Boundary(Delivery, "Command Delivery") {
    Component(grpcApi, "Internal gRPC API", "Go / gRPC", "Authenticates callers; CreateTask for connections/devices/commands; lists opaque candidates and creates video sessions")
    Component(taskService, "Delivery Controller", "Go", "StartDelivery: validates capabilities/target, deduplicates; resolves integration group; atomically journals acceptance and enqueues task for (device_id, group); separate synchronous candidate-list and video use cases")
    Component(consumer, "Task Consumer", "Go / YDB SDK", "Consumes task IDs; submits to worker pool; acknowledges after durable terminal/retry outcome; duplicates handled by versioned group-runtime claim")
    Component(recovery, "Task Recovery Scheduler", "Go", "Scans bounded due retries and expired leases; submits task IDs to common worker pool according to free capacity; does not execute provider calls")
    Component(worker, "Delivery Worker Pool", "Go", "Claims group runtime and task lease; one in-flight command per (device_id, group); appends versioned events and updates projections with fencing; retry/UNKNOWN holds group until reconciled; no heartbeat")
    Component(taskRepo, "Task Repository", "Go / YDB SDK", "Append-only task/group events and current states with independent versions; FIFO pending tasks per (device_id, group); task intents, leases/generations, retries/deadlines and stable provider key")
    Component(gateway, "Provider Adapters", "Go", "Integrate provider APIs: use supplied connection/target context, enforce shared Redis RPS quotas, apply timeouts and normalize results; retry outcome returns to Worker")
    Component(integrationRepo, "Integration Repository", "Go / YDB SDK", "Integration mappings from canonical commands to provider requests and conflict groups; versioned capabilities incl. video; connections, secret refs, candidates, external IDs and tombstones; binding writes check current task lease and operation version atomically")
    Component(publisher, "Delivery Topic Client", "Go / YDB SDK", "Enqueues task IDs atomically with task creation; publishes global result atomically with terminal state; no external IDs in events")
}

ContainerQueue(tasks, "Delivery Task Queue", "YDB Topics", "Private task IDs; task and queue message committed before CreateTask returns")
ContainerDb(database, "Delivery Tables", "YDB", "Task/group event journals, versioned states, per-device/group runtimes and integration data; same database as topics")
ContainerDb(rateLimiter, "Delivery Rate Limiter", "Redis", "Shared RPS quota state across Delivery replicas, scoped by integration/account")
ContainerQueue(results, "Delivery Task Results", "YDB Topics", "Correlated outcomes; verification returns canonical device_type_id; readings carry metric IDs, values, units and timestamps; errors explicit")
System_Ext(providers, "Provider APIs", "Verification, device commands, state and video sessions")

Rel_D(Automation, grpcApi, "CreateTask; idempotent retries", "gRPC / Protobuf")
Rel_D(Monitoring, grpcApi, "CreateTask(ReadState); idempotent retries", "gRPC / Protobuf")
Rel_D(Management, grpcApi, "CreateTask; receives durable task_id", "gRPC / Protobuf")
Rel_D(BFF, grpcApi, "CreateTask(CreateConnection/commands); ListCandidates; CreateVideoSession", "gRPC / Protobuf")
Rel_D(grpcApi, taskService, "StartDelivery / ListCandidates / CreateVideoSession", "Go calls")
Rel_D(taskService, publisher, "Enqueues atomically with task", "Go calls / transaction")
Rel_D(publisher, tasks, "Enqueues task IDs", "YDB SDK / transaction")
Rel_R(consumer, tasks, "Consumes; acknowledges durable processing", "YDB SDK")
Rel_D(consumer, worker, "Submits task_id; waits for durable completion/retry before ack", "Go calls / bounded pool")
Rel_D(worker, taskRepo, "Claims group runtime + lease with expected versions; journals intent before call; fences state writes; releases group only after confirmed resolution", "Go calls / transaction")
Rel_D(recovery, taskRepo, "Finds eligible group tasks and expired leases; reconciles uncertain attempts before redispatch", "Go calls")
Rel_D(recovery, worker, "Submits task IDs up to free capacity", "Go calls / bounded pool")
Rel_D(taskService, integrationRepo, "Checks capabilities/target; resolves command mapping/group; stages binding; loads context for synchronous calls", "Go calls / transaction")
Rel_D(taskService, taskRepo, "Atomically deduplicates, stores resolved group/mapping version, appends accepted event and updates versioned projections", "Go calls / transaction")
worker -[hidden]down-> taskRepo
Rel_D(worker, gateway, "Executes one operation outside DB transaction", "Go calls")
Rel_D(taskService, gateway, "Lists candidates / creates video session with resolved context", "Go calls")
Rel_D(worker, publisher, "Publishes final outcome atomically with task completion", "Go calls / transaction")
Rel_D(integrationRepo, database, "Reads/writes integration data", "YDB SDK")
Rel_D(worker, integrationRepo, "Reads fresh context; applies idempotent binding/deactivation only for current operation and lease generation", "Go calls / transaction")
Rel_D(gateway, providers, "Calls provider after quota acquisition; applies timeout", "HTTPS / JSON")
Rel_D(gateway, rateLimiter, "Atomically acquires shared quota before every outbound request", "Redis protocol")
Rel_D(taskRepo, database, "Append-only events and current task/group states", "YDB SDK")
Rel_D(publisher, results, "Publishes correlated task outcomes", "YDB SDK / transaction")


recovery -[hidden]right-> grpcApi
grpcApi -[hidden]right-> consumer
worker -[hidden]right-> taskService
taskRepo -[hidden]right-> integrationRepo
integrationRepo -[hidden]right-> publisher
publisher -[hidden]right-> gateway
database -[hidden]right-> results
results -[hidden]right-> tasks
tasks -[hidden]right-> providers
providers -[hidden]right-> rateLimiter
@enduml
```

### Monitoring

```puml
@startuml
title WarmHouse - Monitoring Components
!include <C4/C4_Component>
top to bottom direction

ContainerQueue(incoming, "Lifecycle and Delivery Results", "YDB Topics", "DeviceCreated/DeviceDeleted from Management; correlated ReadState results from Delivery")
Container_Boundary(Monitoring, "Monitoring") {
    Component(consumer, "Event Consumer", "Go / YDB SDK", "Consumes lifecycle and Delivery read results; passes offsets to transactional handlers")
    Component(controller, "Monitoring Controller", "Go", "Resolves type profiles; creates/disables schedules; validates metrics; applies availability/freshness rules; ignores stale results")
    Component(scheduler, "Polling Scheduler", "Go", "Periodically invokes Delivery Controller with batch limit; requests freshness checks from Monitoring Controller; no provider calls")
    Component(configRepo, "Monitoring Repository", "Go / YDB SDK", "Type profiles and device schedules; pending poll command IDs, deadlines, next_retry_at, lease owner/generation/expiry and outcomes; deletion tombstones")
    Component(metricsRepo, "Metrics Repository", "Go / YDB SDK", "Observations, latest state, availability and timestamps; conditional updates prevent stale overwrites")
    Component(deliveryController, "Delivery Controller", "Go", "Claims due/retry polls with leases; persists stable command IDs before sending; calls client with bounded concurrency; fences acceptance/error writes; recovers retries without overwriting early final results")
    Component(deliveryClient, "Delivery Client", "Go / gRPC", "CreateTask with supplied command_id, device_id and fixed ReadState command; returns task_id or error; no repository, queue or retry policy")
    Component(publisher, "Telemetry Publisher", "Go / YDB SDK", "Publishes telemetry atomically with observation changes")
}
ContainerDb(database, "Monitoring Tables", "YDB", "Type profiles, device schedules, poll correlations and observations; same database as topics")
Container(Delivery, "Command Delivery", "Go", "CreateTask durably enqueues ReadState; execution results arrive asynchronously")
ContainerQueue(telemetry, "Telemetry Events", "YDB Topic", "BFF and Automation consume state and availability changes")

Rel_U(consumer, incoming, "Consumes lifecycle and read results", "YDB SDK / Protobuf")
Rel_D(consumer, controller, "Updates state + commits offsets atomically", "Go calls / transaction")
Rel_D(controller, configRepo, "DeviceCreated: schedule now; DeviceDeleted: disable and invalidate pending polls; ignore obsolete events/results", "Go calls / transaction")
Rel_D(controller, metricsRepo, "Stores fresh readings and availability", "Go calls / transaction")
Rel_D(controller, publisher, "Publishes changes in observation transaction", "Go calls / transaction")
Rel_D(deliveryController, configRepo, "Claims active schedules up to capacity; persists/reuses command IDs and leases; fences acceptance/errors by generation and lifecycle; preserves terminal outcomes", "Go calls / transaction")
Rel_D(scheduler, deliveryController, "ProcessDue(now, limit); resumes due retries and expired leases", "Go calls")
Rel_D(deliveryController, deliveryClient, "Submits persisted ReadState intent; receives acceptance/error", "Go calls / bounded concurrency; outside DB transaction")
Rel_L(deliveryController, controller, "Reports terminal submission errors/expiry", "Go calls")
Rel_D(scheduler, controller, "Requests freshness/availability checks", "Go calls")
Rel_D(deliveryClient, Delivery, "CreateTask(command_id, device_id, ReadState); receives task_id", "gRPC / Protobuf")
Rel_D(configRepo, database, "Reads type profiles, metric definitions and retry policies; writes schedules and lifecycle projection", "YDB SDK")
Rel_D(metricsRepo, database, "Reads/writes observations", "YDB SDK")
Rel_D(publisher, telemetry, "Publishes state and availability changes", "YDB SDK / Protobuf")

' Align controllers horizontally and keep infrastructure components in one row.
controller -[hidden]right-> deliveryController
publisher -[hidden]right-> metricsRepo
metricsRepo -[hidden]right-> configRepo
configRepo -[hidden]right-> deliveryClient

@enduml
```

Polling Scheduler только инициирует обработку очередной пачки. Delivery Controller выбирает активные расписания, сохраняет намерение опроса и lease до RPC, вызывает Delivery Client вне транзакции и сохраняет принятие/ошибку с проверкой поколения lease и актуальности расписания. Повтор использует прежний `command_id`; ранний финальный результат не перезаписывается ответом о принятии. Delivery Client — простой gRPC-клиент: отправляет фиксированную команду `ReadState` и возвращает `task_id` или ошибку, без собственной политики повторов и хранения состояния.

Monitoring Controller обрабатывает события создания/удаления устройств и результаты опросов. Отдельной проекции устройств нет: версия lifecycle и отметка удаления хранятся в записи расписания для защиты от старых событий. Удаление отключает расписание и инвалидирует ожидаемые опросы; поздний результат не возобновляет опрос и не меняет текущие показания удалённого устройства. Таймаут чтения обрабатывается как недоступность с повторами, а не как удаление устройства. Группу исполнения и интеграционный маппинг команды определяет Delivery.


**Backend For Frontend**

```puml
@startuml
title WarmHouse - BFF Components
!include <C4/C4_Component>
top to bottom direction

Container(WebUI, "Web UI", "JavaScript", "Connects accounts, selects opaque device candidates, manages devices/scenarios and sends commands")
ContainerQueue(events, "Domain Events", "YDB Topics", "Correlated operation success/errors, device lifecycle, telemetry and scenario/run events")
Container_Boundary(BFF, "Backend For Frontend") {
 Component(api, "HTTP API", "Go / HTTP", "Returns operation_id and pending UI card after durable acceptance; exposes operation and object status")
 Component(auth, "Identity and Access", "Go", "Validates token; checks ownership of connections, devices, scenarios and every referenced device")
 Component(controller, "UI Operations Controller", "Go", "Manages create/update/delete operations; persists provisional cards; binds service IDs; applies success/errors; validating only on explicit service status")
 Component(dispatcher, "Operation Dispatcher", "Go", "Claims bounded pending batches with lease/generation; retries stable operation IDs; recovers expired leases; records per-item replies without regressing terminal state")
 Component(client, "Service Clients", "Go / gRPC", "Routes connection/command tasks, candidate lists and video to Delivery; device CRUD to Management; scenario CRUD to Automation")
 Component(consumer, "Projection Consumer", "Go / YDB SDK", "Consumes domain events; delegates projection updates to controller; commits offsets with state changes")
 Component(repo, "BFF Repository", "Go / YDB SDK", "Operation owners/payloads, idempotency keys, leases/retries and outcomes; provisional UI cards by operation_id, final ownership mappings, object versions and tombstones; no external device IDs")
}

Container(Auth, "Auth Service", "Go", "Validates sessions and returns user_id")
Container(Delivery, "Command Delivery", "Go", "CreateTask; scoped opaque candidates; synchronous video sessions")
Container(Management, "Device Management", "Go", "Idempotent create/delete by operation_id; allocates device IDs; reports success and failure")
Container(Automation, "Automation Service", "Go", "Idempotent mutations by operation_id; returns/publishes correlated object ID, version and outcome")
ContainerDb(database, "BFF Tables", "YDB", "Ownership, durable operations and projections")

Rel_D(WebUI, api, "Calls application API", "HTTPS / JSON")
Rel_D(api, auth, "Authenticates and authorizes request", "Go calls")
Rel_L(auth, Auth, "Validates token; receives user_id", "gRPC / Protobuf")
Rel_D(auth, repo, "Checks ownership and ready/active state", "Go calls")
Rel_D(api, controller, "Dispatches after access check", "Go calls")
Rel_D(controller, repo, "Atomically stores operation owner + intent + pending card; reads user-scoped projections", "Go calls / transaction")
Rel_D(controller, client, "Lists candidates or creates video session after ownership check", "Go calls")
Rel_D(dispatcher, repo, "Claims limited batches; fences acceptance/rejection and retries by lease generation; correlates returned IDs", "Go calls / transaction")
Rel_D(dispatcher, client, "Retries mutations outside DB transaction", "Go calls")
Rel_D(client, Delivery, "CreateTask(CreateConnection/command); ListCandidates(connection_id); CreateVideoSession(device_id)", "gRPC / Protobuf")
Rel_D(client, Management, "CreateDevice(connection_id, candidate_id, operation_id); DeleteDevice", "gRPC / Protobuf")
Rel_D(client, Automation, "Scenario mutations with stable operation_id and expected version after ownership checks", "gRPC / Protobuf")
Rel_U(consumer, events, "Consumes outcomes and state events", "YDB SDK / Protobuf")
Rel_D(consumer, controller, "Applies correlated event with transaction and offset", "Go calls / transaction")
Rel_D(controller, repo, "Deduplicates events; binds returned object ID to operation owner; applies versioned success/error + offset without state regression", "Go calls / transaction")
Rel_D(repo, database, "Reads/writes BFF-owned data", "YDB SDK")

@enduml
```

**Automation Service**

```puml
@startuml
title WarmHouse - Automation Service Components
!include <C4/C4_Component>
top to bottom direction

Container(BFF, "BFF", "Go", "Checks scenario and referenced device ownership before every mutation")
ContainerQueue(incoming, "Automation Input Events", "YDB Topics", "Telemetry and Delivery task results")

Container(DeviceManagement, "Device Management", "Go", "Owns device lifecycle and effective capabilities; publishes versioned snapshots/tombstones")
ContainerQueue(deviceEvents, "Device Lifecycle Events", "YDB Topics", "DeviceCreated / DeviceUpdated / DeviceDeleted; Automation subscription")

Container_Boundary(Automation, "Automation Service") {
 together {
 Component(lifecycleConsumer, "Device Projection Consumer", "Go / YDB SDK", "Applies DeviceCreated/DeviceUpdated snapshots and DeviceDeleted tombstones by source version; deduplicates and commits offsets atomically; requests affected scenario revalidation")
 Component(api, "Scenario gRPC API", "Go / gRPC", "Authenticates BFF; validates request format; delegates scenario mutations to Controller")
 }
 Component(consumer, "Execution Event Consumer", "Go / YDB SDK", "Batches private run requests and Delivery results to Orchestrator; acknowledges after durable handling")
 Component(triggerConsumer, "Trigger Event Consumer", "Go / YDB SDK", "Consumes telemetry only; forwards to Trigger Evaluator; independent subscription and offsets")
 Component(controller, "Scenario Controller", "Go", "Deduplicates mutations by operation_id; validates commands/parameters against device projection and expected scenario version; revalidates affected definitions; marks incompatible scenarios unavailable; rejects definition edits while runtime exists")
 Component(evaluator, "Trigger Evaluator", "Go", "Updates telemetry projection; evaluates enabled triggers only for locally valid scenarios; atomically starts runtime only if absent; journals/enqueues deduplicated RunRequested")
 Component(engine, "Scenario Orchestrator", "Go", "Resumes scenario runtime with same run_id; clears runtime on completion; checks device capabilities before each action; dispatches one stage; recovers due retries/expired leases; correlates Delivery results")
 Component(scheduler, "Scenario Scheduler", "Go", "Batches due schedules; checks device projection, absent runtime and expected version; atomically starts runtime, journals/enqueues RunRequested and advances next_run_at; stable occurrence IDs")
 Component(dispatcher, "Delivery Dispatcher", "Go / gRPC", "Calls CreateTask with supplied stable command_id and timeout; returns task_id or error to Orchestrator; no queue, repository or event publication")
 Component(repo, "Scenario Repository", "Go / YDB SDK", "Current scenario state with event version and optional runtime, schedules, telemetry projection, durable revalidation work and append-only scenario event journal; event dedup, stage leases/generations, retries/deadlines and command correlations; no user IDs")
 Component(devices, "Device Projection Repository", "Go / YDB SDK", "ApplySnapshot / MarkDeleted conditionally on source_version; Get / GetMany for scenario validation; accesses automation_device_projection")
 Component(publisher, "Automation Topic Client", "Go / YDB SDK", "Transports private run requests; publishes definition/runtime snapshots and run outcomes atomically with owning state changes")
}

Container(Delivery, "Command Delivery", "Go", "Rechecks current target/capabilities; resolves provider mappings and conflict groups; accepts tasks idempotently and publishes outcomes")
ContainerQueue(internal, "Automation Internal Events", "YDB Topics", "Private RunRequested with message_id, scenario_id and runtime run_id")
ContainerDb(database, "Automation Tables", "YDB", "Scenario state/runtime, telemetry projection, revalidation work, schedules and scenario event journal; same database as topics")
ContainerDb(deviceProjection, "automation_device_projection", "YDB table; same Automation database", "PK device_id; device_type_id, lifecycle, supported_commands, parameter_constraints, source_version; deleted lifecycle retained as tombstone")
ContainerQueue(events, "Scenario and Run Events", "YDB Topics", "BFF projects scenario status and shows execution errors to the user")

Rel_R(publisher, internal, "Publishes run requests atomically", "YDB SDK")
Rel_R(consumer, internal, "Consumes private events", "YDB SDK")
Rel_D(BFF, api, "Manages authorized scenarios", "gRPC / Protobuf")
Rel_U(consumer, incoming, "Consumes Delivery task results only", "YDB SDK / Protobuf")
Rel_U(triggerConsumer, incoming, "Consumes telemetry only", "YDB SDK / Protobuf")
Rel_D(DeviceManagement, deviceEvents, "Publishes committed device snapshots and deletion tombstones", "YDB SDK / Protobuf")
Rel_U(lifecycleConsumer, deviceEvents, "Consumes versioned device events; commits offset with projection update", "YDB SDK / Protobuf")
Rel_D(api, controller, "Creates, updates, deletes and enables/disables scenarios", "Go calls")
Rel_D(consumer, engine, "Forwards RunRequested and Delivery results", "Go calls / transaction")
Rel_D(triggerConsumer, evaluator, "Forwards telemetry", "Go calls / transaction")
Rel_D(lifecycleConsumer, devices, "DeviceCreated/Updated: ApplySnapshot; DeviceDeleted: MarkDeleted; ignore older versions", "Go calls / transaction")
Rel_D(lifecycleConsumer, repo, "Durably schedules affected scenario revalidation with projection update", "Go calls / transaction")
Rel_D(lifecycleConsumer, controller, "Requests revalidation; durable work resumes after restart", "Go calls")
Rel_D(controller, devices, "Checks lifecycle, supported commands and parameter constraints", "Go calls / shared transaction")
Rel_D(evaluator, devices, "Checks referenced devices before creating runtime", "Go calls / shared transaction")
Rel_D(scheduler, devices, "Checks referenced devices before creating runtime", "Go calls / shared transaction")
Rel_D(engine, devices, "Checks action target/command before persisting dispatch intent", "Go calls / shared transaction")
Rel_D(devices, deviceProjection, "Conditional UPSERT by source_version; retains deletion tombstones; SELECT for checks", "YDB SDK / shared transaction")
Rel_D(controller, repo, "Persists mutation outcome + current definition with expected event version and operation_id; processes durable revalidation; records invalid definitions without editing active plans", "Go calls / transaction")
Rel_D(controller, publisher, "Publishes definition/lifecycle outcomes with operation_id, object ID/version; unavailable status with state", "Go calls / transaction")
Rel_D(evaluator, repo, "Updates projection/trigger state; deduplicates; checks absent runtime + expected version; creates runtime + journals RunRequested + commits input offsets", "Go calls / transaction")
Rel_D(evaluator, publisher, "Enqueues RunRequested with trigger state and telemetry offsets", "Go calls / transaction")
Rel_D(scheduler, repo, "Checks absent runtime + expected version; creates runtime, advances due schedule and appends RunRequested with occurrence_id", "Go calls / transaction")
Rel_D(scheduler, publisher, "Enqueues RunRequested in same schedule/journal transaction", "Go calls / transaction")
Rel_D(engine, repo, "Deduplicates events; validates runtime run_id, expected version and stage/generation; atomically appends journal, updates runtime + offsets; claims leases and recovers due stages", "Go calls / transaction")
Rel_D(engine, publisher, "Publishes terminal run outcomes with journal/state", "Go calls / transaction")
Rel_D(engine, dispatcher, "Calls CreateTask for leased action; receives task_id/error; bounded concurrency", "Go calls / outside DB transaction")
Rel_D(dispatcher, Delivery, "CreateTask(command_id, device_id, command); retry same key", "gRPC / Protobuf")
Rel_D(repo, database, "Reads/writes scenario state, telemetry and revalidation work", "YDB SDK")
Rel_D(publisher, events, "Publishes scenario and run status", "YDB SDK / Protobuf")

triggerConsumer -[hidden]right-> lifecycleConsumer
api -[hidden]left-> lifecycleConsumer
evaluator -[hidden]right-> controller
controller -[hidden]right-> scheduler
@enduml
```

Проекция устройств в Automation отделена от проекции телеметрии для триггеров. Таблица `automation_device_projection` явно показана на L3 как локальная таблица Automation в той же YDB database, а не отдельная база или сервис. Её обновляет Device Projection Consumer через Device Projection Repository: создание/обновление вызывает `ApplySnapshot`, удаление — `MarkDeleted`; запись принимается только при более новой `source_version`. Consumer фиксирует обновление, задание перепроверки и offset одной транзакцией. Controller, Trigger Evaluator, Scheduler и Orchestrator читают таблицу через этот repository. Device Management владеет общими определениями типов и команд и публикует полные снимки `DeviceCreated`/`DeviceUpdated`: `device_id`, `device_type_id`, состояние жизненного цикла, фактически поддерживаемые общие команды, ограничения параметров и монотонную версию устройства. Изменение общих определений, влияющее на возможности устройства, приводит к публикации нового снимка устройства. `DeviceDeleted` сохраняется как tombstone: позднее событие не восстанавливает удалённое устройство. Версия источника не смешивается с событийной версией сценария. Группы конфликтов, внешние идентификаторы и преобразования API провайдера остаются в Delivery; владение и проверка пользовательского доступа — в BFF.

Consumer атомарно обновляет проекцию, сохраняет offset и создаёт долговечное задание перепроверки затронутых сценариев. Controller обрабатывает такие задания с восстановлением после перезапуска и помечает несовместимые определения недоступными. Проверка проекции выполняется при сохранении/включении сценария, создании runtime и перед сохранением намерения отправить очередное действие, в общей транзакции с соответствующим изменением состояния сценария. Уже отправленная команда не отменяется изменением проекции; её результат обрабатывается по прежнему `command_id`. Отсутствующий или удалённый объект и неподдерживаемая команда не допускаются к отправке; активный runtime завершается с соответствующей ошибкой, если следующее действие стало недопустимым. Изменение доступности сценария не переписывает его выполняемое определение.

Проекция является eventual consistent: Delivery проверяет актуальные возможности повторно перед исполнением. Первоначальное заполнение и восстановление требуют согласованного снимка Device Management с позицией продолжения потока либо полного воспроизведения сохранённых событий; до заполнения необходимой записи действия для устройства не допускаются. Эти правила относятся к проектируемому контракту событий, а не к уже реализованному API.

Delivery Dispatcher — обычный синхронный вызов CreateTask с таймаутом: возвращает task_id или ошибку. Он не читает и не пишет очереди. Orchestrator сохраняет намерение и lease до вызова, а принятие/ошибку после вызова — с проверкой lease generation. Восстановление использует тот же command_id; ранний итог Delivery не перезаписывается подтверждением принятия. Приватный топик Automation переносит RunRequested, журнал сохраняет все переходы state machine.

**Диаграмма кода (Code)**

### Command Delivery

Проектируемая структура Go-кода (C4 L4), а не описание уже реализованного сервиса. Сигнатуры сокращены: `context.Context`, типы транспортных DTO и детали реализации опущены. Зависимости показаны стрелками либо типами полей структур.

```puml
@startuml
title WarmHouse - Command Delivery Code (Go design)
top to bottom direction
hide empty members
class DeliveryController <<struct>> {
 - tasks: TaskRepository
 - integrations: IntegrationRepository
 - topics: DeliveryTopicClient
 - tx: UnitOfWork
 + StartDelivery(request): TaskID, error
 + ListCandidates(connectionID): []Candidate, error
 + CreateVideoSession(deviceID): VideoSession, error
}
class TaskConsumer <<struct>> {
 - workers: DeliveryWorkerPool
 + Handle(taskID, offset): error
}
class TaskRecoveryScheduler <<struct>> {
 - tasks: TaskRepository
 - workers: DeliveryWorkerPool
 + RecoverDue(now, limit): error
}
class DeliveryWorkerPool <<struct>> {
 - capacity: int
 - tasks: TaskRepository
 - integrations: IntegrationRepository
 - providers: ProviderAdapters
 - topics: DeliveryTopicClient
 - tx: UnitOfWork
 + Submit(taskID): Completion, error
 - execute(task, lease): error
 - persistOutcome(task, lease, outcome): error
}
class Task <<struct>> {
 + ID, CallerID, IdempotencyKey: string
 + DeviceID, Kind: string
 + Version: uint64
 + GroupKey: DeliveryGroupKey
 + MappingVersion: uint64
 + Command: Command
 + Status: pending | executing | retry | succeeded | failed | unknown
 + AttemptIntent: Attempt
 + NextRetryAt, Deadline: time.Time
 + ProviderIdempotencyKey: string
}
class DeliveryGroupKey <<struct>> {
 + DeviceID, Group: string
}
class DeliveryGroupState <<struct>> {
 + Key: DeliveryGroupKey
 + Version: uint64
 + Runtime: *DeliveryRuntime
}
class DeliveryRuntime <<struct>> {
 + TaskID: string
 + Lease: Lease
}
class Lease <<struct>> {
 + TaskID, Owner: string
 + GroupKey: DeliveryGroupKey
 + Generation: uint64
 + ExpiresAt: time.Time
}
interface TaskRepository {
 + CreateOrGet(tx, request, mapping): Task, bool, error
 + FindDue(now, limit): []TaskID, error
 + ClaimNext(tx, taskID, expectedTaskVersion, expectedGroupVersion, owner, ttl): Task, Lease, error
 + SaveAttempt(tx, lease, expectedTaskVersion, expectedGroupVersion, intent): error
 + ScheduleRetry(tx, lease, expectedTaskVersion, expectedGroupVersion, retryAt, reason): error
 + Finish(tx, lease, expectedTaskVersion, expectedGroupVersion, outcome): error
}
interface IntegrationRepository {
 + ValidateAndStage(tx, request): error
 + ResolveMapping(tx, deviceID, commandCode): CommandMapping, error
 + ResolveCurrent(task): ProviderContext, error
 + ApplyVerifiedResult(tx, lease, result): error
}
class CommandMapping <<struct>> {
 + ProviderType, DeviceTypeID, CommandCode: string
 + Group: string
 + Version: uint64
 + RequestMapping, ResponseMapping: Mapping
}
class ProviderContext <<struct>> {
 + ProviderType, ConnectionID: string
 + ExternalDeviceID: string
 + CredentialRef, QuotaKey: string
 + ConfigurationVersion: uint64
 + Mapping: CommandMapping
 + Deadline: time.Time
}
class ProviderAdapters <<struct>> {
 - adapters: map[ProviderType]ProviderAdapter
 + Execute(command, context, key): Result, error
 + Reconcile(attempt, context): Outcome, error
 + ListCandidates(context): []Candidate, error
 + CreateVideoSession(context): VideoSession, error
}
interface ProviderAdapter {
 + Execute(command, context, key): Result, error
 + Reconcile(attempt, context): Outcome, error
}
class RedisRateLimiter <<struct>> {
 + Acquire(quotaKey, deadline): error
}
interface DeliveryTopicClient {
 + EnqueueTask(tx, taskID): error
 + PublishGlobalResult(tx, result): error
 + Acknowledge(offset): error
}
interface UnitOfWork {
 + WithinTx(fn: func(tx) error): error
}
TaskConsumer -down-> DeliveryWorkerPool : task ID; await durable outcome
TaskRecoveryScheduler -down-> DeliveryWorkerPool : bounded recovery submissions
TaskRecoveryScheduler -down-> TaskRepository : FindDue; claim remains in worker
DeliveryWorkerPool -down-> TaskRepository : claim group runtime; journal intent; fenced transitions
DeliveryWorkerPool -down-> IntegrationRepository : fresh context; verified effects in final transaction
DeliveryWorkerPool -down-> ProviderAdapters : bounded external call; reconcile uncertain attempt
DeliveryWorkerPool -down-> DeliveryTopicClient : final event in task completion transaction
TaskConsumer -left-> DeliveryTopicClient : ack after durable outcome
DeliveryController -down-> TaskRepository : deduplicate; enqueue within resolved device/group
DeliveryController -down-> DeliveryTopicClient : atomic task enqueue
DeliveryController -down-> IntegrationRepository : validate/stage; resolve command group
DeliveryController -down-> ProviderAdapters : synchronous list/video
IntegrationRepository -down-> ProviderContext
ProviderAdapters -down-> ProviderAdapter
ProviderAdapter -down-> RedisRateLimiter : quota before each request
DeliveryGroupState -down-> DeliveryRuntime
DeliveryRuntime -down-> Lease
TaskRepository -down-> Task
TaskRepository -down-> DeliveryGroupState
DeliveryGroupState -down-> DeliveryGroupKey
IntegrationRepository -down-> CommandMapping
note right of TaskRepository
 Device-command state is unique by (device_id, group).
 CreateOrGet deduplicates by caller/key; stores resolved group and mapping version.
 Pending tasks retain acceptance order within their group.
 ClaimNext starts only the oldest eligible task when group runtime is absent;
 recovery resumes its existing task, never replaces an active runtime.
 Mutations append events and update current state with expected versions.
 Lease generation fences stale writes within the group.
 Retry/UNKNOWN keeps runtime; expiry requires reconciliation.
 Finish clears runtime only after confirmed resolution; other groups run independently.
end note
note right of CommandMapping
 Integration resolves group from device and canonical command.
 Callers do not supply group; retries reuse the stored mapping.
 Group changes require draining conflicting accepted tasks before switching.
 Connection-only lifecycle tasks retain their separate operation scope.
end note
DeliveryTopicClient -[hidden]down-> UnitOfWork
@enduml
```

# Задание 3. Разработка ER-диаграммы

Диаграмма отражает ключевые сущности целевой системы, основные атрибуты и кратность связей. Технические сущности обработки задач и хранения проекций опущены.

```puml
@startuml
title WarmHouse — ключевые сущности
hide circle
skinparam linetype polyline
skinparam nodesep 35
skinparam ranksep 45
skinparam shadowing false
skinparam packageStyle rectangle

package "Устройства и доступ" {
  entity "Пользователь" as user {
   * user_id <<PK>>
   --
   login <<UQ>>
   password_hash
  }
  entity "Провайдер" as provider {
   * provider_id <<PK>>
   --
   name
   api_configuration
  }
  entity "Подключение к провайдеру" as connection {
   * connection_id <<PK>>
   --
   provider_id <<REF>>
   secret_ref
   status
  }
  entity "Устройство" as device {
   * device_id <<PK>>
   --
   device_type_id <<REF>>
   name
   status
  }
  entity "Показание устройства" as reading {
   * observation_id <<PK>>
   --
   device_id <<REF>>
   metric_code
   value
   unit
   measured_at
  }
}

package "Каталог команд" {
  entity "Тип устройства" as deviceType {
   * device_type_id <<PK>>
   --
   name
  }
  entity "Общая команда" as commandDefinition {
   * command_code <<PK>>
   --
   name
   parameter_schema
  }
  entity "Команда типа устройства" as supportedCommand {
   * device_type_id <<PK, REF>>
   * command_code <<PK, REF>>
   --
   parameter_constraints
  }
  entity "Интеграционный маппинг команды" as commandMapping {
   * provider_id <<PK, REF>>
   * device_type_id <<PK, REF>>
   * command_code <<PK, REF>>
   --
   group
   request_mapping
   response_mapping
  }
}

package "Автоматизация" {
  entity "Сценарий" as scenario {
   * scenario_id <<PK>>
   --
   name
   version
   runtime <<optional>>
   enabled
   trigger_condition
   schedule
  }
  entity "Действие сценария" as action {
   * scenario_id <<PK, REF>>
   * step_id <<PK>>
   --
   device_id <<REF>>
   command_code <<REF>>
   parameters
  }
}

package "Исполнение" {
  entity "Состояние доставки группы" as deliveryState {
   * device_id <<PK, REF>>
   * group <<PK>>
   --
   version
   runtime <<optional>>
  }
  entity "Команда устройству" as command {
   * command_id <<PK>>
   --
   device_id <<REF>>
   scenario_id <<REF, optional>>
   run_id <<optional>>
   step_id <<optional>>
   version
   group <<REF>>
   command_code <<REF>>
   parameters
   status
   result
  }
}

user ||--o{ connection : владеет
user ||--o{ device : владеет
user ||--o{ scenario : владеет
provider ||--o{ connection : подключения
connection ||--o{ device : доступ через API
deviceType ||--o{ device : определяет тип
deviceType ||--o{ supportedCommand : поддерживает
commandDefinition ||--o{ supportedCommand : доступна типам
supportedCommand ||--o{ commandMapping : реализуется интеграцией
scenario ||--o{ action : содержит
device ||--o{ action : используется в
device ||--o{ deliveryState : группы доставки
deliveryState ||--o{ command : исполняет последовательно
device ||--o{ reading : измерения

legend bottom
 PK — ключ; UQ — уникальное значение; REF — логическая ссылка.
 || — один; o{ — ноль или много.
 Между блоками показаны основные связи; остальные ссылки заданы полями:
 action.command_code и command.command_code → Общая команда.command_code.
 commandMapping.provider_id → Провайдер.provider_id.
 command.scenario_id → Сценарий.scenario_id (необязательно).
 command.(device_id, group) → Состояние доставки группы.(device_id, group).
 У каждого сценария один пользователь-владелец; связь хранится в BFF.
 version — текущая событийная версия состояния.
 runtime — активный запуск сценария либо текущая команда группы.
 group определяется интеграцией в Delivery; разные группы могут работать параллельно.
 Журналы событий и технические проекции опущены; связи не означают физические FK.
endlegend
@enduml
```

Общие определения команд, схемы параметров и поддержка команд типами устройств принадлежат Device Management. Ограничения типа уточняют общую схему параметров. Delivery владеет интеграционными маппингами: для сочетания провайдера, типа устройства и общей команды задаёт преобразование запроса/ответа и группу конфликтов. Automation хранит в действии `device_id`, `command_code` и параметры; группу не передаёт. Delivery определяет её при принятии команды и сохраняет в запросе. Изменение маппинга не должно разделять старые и новые конфликтующие запросы на независимые группы: изменение групп требует согласованного переключения после завершения уже принятых конфликтующих команд.

Состояние доставки хранится по ключу `(device_id, group)` и обновляется с проверкой событийной `version`. Его `runtime` содержит текущую команду. Каждый запрос сохраняет собственный `command_id` для дедупликации и корреляции; ожидающие запросы не перезаписывают runtime занятой группы. Очерёдность внутри группы сохраняется, но технические позиции очереди на обзорной ER не показаны. При неизвестном исходе текущей команды группа остаётся занятой до сверки. Параллелизм между группами допустим в пределах ограничений интеграции.

`runtime` сценария содержит единственный активный запуск и очищается при завершении. Поля `scenario_id`, `run_id` и `step_id` у сценарной команды сохраняют связь с её происхождением после очистки runtime; у самостоятельного запроса они отсутствуют. Версии сценария, состояния доставки и отдельного запроса независимы. Append-only являются журналы событий, а не записи текущего состояния. Владение пользователем хранится в BFF, привязка устройства к подключению и внешний идентификатор — в Delivery; связи на диаграмме логические. Технические проекции Automation на этой ER не показаны.



# Задание 4. Создание и документирование API

### 1. Тип API

В целевой системе используются три вида взаимодействия:

- **HTTPS / JSON для Web UI → Auth и BFF.** Пользовательские операции описаны в OpenAPI: регистрация и сессии, подключения, устройства, сценарии, команды и просмотр состояния. BFF проверяет владение объектами и всеми устройствами, указанными в сценарии.
- **gRPC / Protobuf для синхронных вызовов между сервисами.** Типизированные контракты описывают проверку токена, изменение устройств/сценариев, принятие задачи Delivery, получение кандидатов и создание видеосессии. Ответ о принятии задачи не означает, что устройство уже выполнило команду.
- **Асинхронные события через YDB Topics.** Результаты команд, изменения устройств и сценариев, показания доставляются подписчикам независимо от исходного запроса. Контракты событий описаны в AsyncAPI, формат передачи — Protobuf. Monitoring получает события и отправляет фиксированную команду `read_state`; отдельного входящего прикладного RPC у него нет.

Такое разделение позволяет завершать HTTP-запрос после долговечного принятия операции, не ожидая внешнего провайдера. BFF возвращает `202` и `operation_id`; пользователь получает результат через статус операции и проекции объектов. Повтор запроса использует прежний ключ идемпотентности. Конкурентные изменения проверяют ожидаемую событийную версию. Delivery самостоятельно определяет группу команды по интеграционному маппингу и последовательно исполняет запросы внутри `(device_id, group)`.

### 2. Документация API


| Контракт | Документация |
|---|---|
| Auth: регистрация, вход, обновление токена, выход | [OpenAPI Auth](api/auth.openapi.json) |
| BFF: пользовательские операции, состояния объектов и показания | [OpenAPI BFF](api/bff.openapi.json) |
| Auth, Device Management, Automation, Command Delivery: внутренние RPC | [Protobuf services](api/services.proto) |
| События Device Management, Delivery, Monitoring и Automation | [AsyncAPI](api/events.asyncapi.json) |
| Формат событий для YDB Topics | [Protobuf events](api/events.proto) |
| Ответственность сервисов, авторизация, ошибки, версии и примеры взаимодействия | [Описание контрактов](api/README.md) |

Все события используют единый `FactMessage`: идентификатор, времена возникновения/записи, корреляция и причина, необязательные данные сущности и типизированный `oneof payload`. В метаданных сообщения YDB передаются `producer`, `payload_type` и `schema_version`, чтобы подписчики отбрасывали ненужные типы до декодирования тела.

OpenAPI описывает HTTP-интерфейсы; внутренние gRPC-вызовы документированы отдельно в Protobuf. AsyncAPI покрывает событийные интерфейсы микросервисов, включая Monitoring. В спецификациях предусмотрены примеры запросов и событий, состояния `pending/succeeded/failed/unknown` и корреляция `operation_id/command_id/run_id`.

# Задание 5. Работа с docker и docker-compose

Перейдите в apps.

Там находится приложение-монолит для работы с датчиками температуры. В README.md описано как запустить решение.

Вам нужно:

1) сделать простое приложение temperature-api на любом удобном для вас языке программирования, которое при запросе /temperature?location= будет отдавать рандомное значение температуры.

Locations - название комнаты, sensorId - идентификатор названия комнаты

```
	// If no location is provided, use a default based on sensor ID
	if location == "" {
		switch sensorID {
		case "1":
			location = "Living Room"
		case "2":
			location = "Bedroom"
		case "3":
			location = "Kitchen"
		default:
			location = "Unknown"
		}
	}

	// If no sensor ID is provided, generate one based on location
	if sensorID == "" {
		switch location {
		case "Living Room":
			sensorID = "1"
		case "Bedroom":
			sensorID = "2"
		case "Kitchen":
			sensorID = "3"
		default:
			sensorID = "0"
		}
	}
```

2) Приложение следует упаковать в Docker и добавить в docker-compose. Порт по умолчанию должен быть 8081

3) Кроме того для smart_home приложения требуется база данных - добавьте в docker-compose файл настройки для запуска postgres с указанием скрипта инициализации ./smart_home/init.sql

Для проверки можно использовать Postman коллекцию smarthome-api.postman_collection.json и вызвать:

- Create Sensor
- Get All Sensors

Должно при каждом вызове отображаться разное значение температуры

Ревьюер будет проверять точно так же.


