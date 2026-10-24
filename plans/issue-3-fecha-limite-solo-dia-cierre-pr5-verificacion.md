# Feature: Fecha límite opcional solo día en las tareas (issue #3, cierre del PR #5)

## Encargo y criterios de aceptación

Encargo (issue #3 de clerma-arch/desarrollo-agentico-todo-03, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"):
añadir a cada tarea una fecha límite opcional que guarde solo el día, sin hora.

Criterios de aceptación:

1. Crear una tarea sin fecha funciona como antes y la tarea queda sin `due_date` (NULL).
2. Crear una tarea con fecha válida `YYYY-MM-DD` la persiste como día (sin hora).
3. Una fecha con formato inválido no crea la tarea, muestra el mensaje "La fecha no es válida." y conserva el valor introducido.
4. Editar una tarea permite fijar, cambiar y borrar la fecha límite (campo vacío = sin fecha).
5. Editar el título sin tocar la fecha no altera la fecha guardada.
6. La lista muestra "Fecha límite: DD/MM/YYYY" solo cuando la tarea tiene fecha.
7. Completar o reabrir una tarea conserva su fecha límite.

Requisitos explícitos (de la issue): fecha opcional, solo día, sin hora.
Inferencias (no explícitas en la issue): formato `YYYY-MM-DD` del input nativo `type="date"`; mensaje de error en español; la fecha se valida en servidor aunque el navegador ya la limite.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (ya commiteado y empujado; `origin` no tiene commits pendientes; PR #5 abierto, título "feat(tasks): añadir fecha límite opcional (solo día) a las tareas"):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `.down.fizz`.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en `time.Parse`; `TasksCreate` y `TasksUpdate` leen `due_date`, usan `task.Validate(tx)` y añaden el error `due_date` a `verrs`; al fallar re-renderizan con `dueDateInput`.
- Plantillas `templates/tasks/index.plush.html` (input date en el formulario de alta y etiqueta en la lista) y `templates/tasks/edit.plush.html` (campo fecha con ayuda y error).
- `public/assets/app.css`: `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests en `actions/tasks_test.go` (crear con/sin fecha, fecha inválida en alta y edición, conservar fecha al completar, mostrar fecha en lista, prefill en edición, fijar/borrar fecha) y `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`).

Propuesta: no rediseñar. Verificar que el código actual cumple los criterios 1–7, completar los huecos de test que se detecten y ejecutar la validación. No hay cambios de comportamiento previstos.

Pendiente de confirmar al implementar (lectura del código, sin ejecutar aún):

- `Test_TasksUpdate_InvalidDueDate` y `Test_TasksUpdate_SetsAndClearsDueDate` existen; hay que leerlos para confirmar que cubren borrar la fecha (campo vacío) y que no alteran la fecha cuando el campo no se envía. Si el campo no se envía, `input.DueDate` queda vacío y `TasksUpdate` lo interpreta como "sin fecha": **esto borraría la fecha al editar solo el título si el formulario no la envía**. Comprobar en `edit.plush.html` que el input siempre se envía (lo hace, es un input `type="date"` dentro del mismo form), pero documentar el comportamiento: la API con JSON sin `due_date` también borra la fecha.
- `Test_TasksIndex_ShowsDueDate` debe comprobar el formato `DD/MM/YYYY` en la lista.

## Alcance

Incluido:

- Verificación del código y tests existentes contra los criterios.
- Añadir tests solo si un criterio no tiene prueba que lo demuestre.
- Ejecución de la validación (gofmt, vet, test, build) y actualización del PR #5 con push a `origin`.

Excluido:

- Cambios de hora, zona horaria, recordatorios, filtros por fecha o ordenación por fecha.
- Cambios en `migrations/schema.sql` (archivo sin seguimiento generado, fuera del PR).
- Cambios en `config/adw.yml` (modificado y sin relación con la feature).
- Los ~160 archivos `plans/issue-3-*` sin seguimiento: no se añaden al PR.

## Diseño por capas

- Persistencia y migraciones: ya hecho (columna `date` nullable). No se cambia.
- Modelo: ya hecho (`*time.Time`). No se cambia.
- Acción y ruta: ya hecho (`TasksCreate`, `TasksUpdate`, `parseDueDate`). Solo se cambia si un test de hueco revela un fallo.
- Renderizado: ya hecho en Plush. No se cambia salvo fallo detectado.
- CSS: ya hecho. No se cambia.
- JavaScript: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz`: columna `due_date`.
- `models/task.go`: modelo y validación de título.
- `models/task_test.go`: pruebas del modelo para fecha.
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: pruebas de acción para fecha.
- `actions/actions_test.go`: helper `authed` con token CSRF usado por los tests que cambian estado.
- `templates/tasks/index.plush.html`: input de alta y etiqueta en lista.
- `templates/tasks/edit.plush.html`: campo de edición.
- `public/assets/app.css`: estilos `.task-due` y `.new-task-row input[type="date"]`.
- `database.yml`: la base de test es `todo_test`.

## Implementación paso a paso

1. Confirmar que la rama está limpia en lo relevante: `git status --short --branch`. Esperar `feat/issue-3-add-task-due-date...origin/feat/issue-3-add-task-due-date` sin commits pendientes. Los cambios ajenos (`config/adw.yml`, `.claude/skills/ship/`, `migrations/schema.sql`, `plans/`) quedan fuera.
2. Leer `actions/tasks_test.go` (líneas 259–end) y `models/task_test.go` (76–end) y contrastar con los criterios 1–7. Anotar cualquier criterio sin prueba.
3. Para cada criterio sin prueba, añadir un test en el mismo archivo y con el mismo patrón (`as.authed`, `as.DB`, `url.Values`). No cambiar código de aplicación salvo que el test falle por un defecto real; si falla, detenerse y reportarlo.
4. Ejecutar la validación de la sección siguiente, con `GO_ENV=test`.
5. Si todo pasa, hacer commit solo de los archivos de test tocados y del plan (lista explícita), y empujar a `origin` la misma rama. El PR #5 se actualiza solo.

## Tests y validación

Prerrequisitos: PostgreSQL accesible con la base `todo_test` y migraciones aplicadas. Los tests Go requieren `GO_ENV=test`; sin ella la app escribe en `development`.

Comandos, en este orden:

- `gofmt -l .` debe devolver vacío.
- `go vet ./...` debe terminar sin errores.
- `GO_ENV=test go test -p 1 ./...` debe pasar todo, incluidos `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
- `go build ./...` debe compilar.

Criterio de cierre: los cuatro comandos en verde. Si PostgreSQL no está disponible, anotarlo como bloqueo, no marcar la validación como superada.

## Riesgos y recuperación

- PostgreSQL no disponible o sin migraciones de test: los tests de acción fallan al conectar. Recuperación: levantar la base con el skill `env:start` y volver a ejecutar; no simular resultados.
- Un test de hueco revela que editar sin enviar `due_date` borra la fecha. Recuperación: no cambiar el alcance; documentarlo como comportamiento actual y decidir con el usuario si el contrato JSON debe conservar la fecha cuando el campo no llega.
- Zona horaria: `time.Parse` usa UTC y el formato `02/01/2006` de la lista usa la fecha tal cual; con columna `date` no debería desplazar el día. Si un test lo contradice, es un defecto a reportar.
- Los planes sin seguimiento y los archivos ajenos pueden entrar al commit por error: usar siempre `git add` con rutas explícitas.
