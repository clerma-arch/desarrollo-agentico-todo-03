# Feature: Fecha límite opcional por tarea (solo día) — issue #3

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fecha límite a las tareas", cuerpo "vale un dia sin hora".

Requisitos explícitos:

1. Cada tarea puede tener una fecha límite opcional.
2. La fecha es solo día, sin hora (`YYYY-MM-DD`).

Criterios de aceptación:

- AC1. Al crear una tarea, el campo `due_date` es opcional; sin valor, la tarea queda sin fecha.
- AC2. Al crear una tarea con `due_date` válido, la fecha se guarda como día sin hora (hora 00:00).
- AC3. Al editar una tarea se puede poner, cambiar o borrar la fecha límite.
- AC4. El índice muestra "Fecha límite: DD/MM/AAAA" solo en las tareas que tienen fecha.
- AC5. Una fecha no válida (p. ej. `no-es-fecha`) no se guarda: se responde 422 y se muestra el error junto al campo, conservando lo escrito.
- AC6. Un error de título no pierde la fecha introducida en el formulario.
- AC7. Completar o reabrir una tarea conserva su fecha límite.
- AC8. Las tareas existentes (sin fecha) siguen funcionando.

## Comportamiento actual y propuesta

Estado observado en la rama `feat/issue-3-add-task-due-date` (ya tiene commits de implementación, no solo planes):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: añade `due_date` tipo `date`, nullable. Su `down.fizz` la elimina.
- `models/task.go`: campo `DueDate *time.Time` con tags `json`, `db` y `form`.
- `actions/tasks.go`: `parseDueDate` (vacío → nil; formato `2006-01-02`; error → mensaje `dueDateInvalidMessage`). Alta y actualización leen `due_date`, validan y re-renderizan con 422 si hay errores. `TasksEdit` y `TasksIndex` inicializan `dueDateInput`.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta, muestra `.task-due` con la fecha formateada, y los errores de `due_date`.
- `templates/tasks/edit.plush.html`: campo `Fecha límite` con ayuda y errores.
- `public/assets/app.css`: estilos `.task-due` y `.new-task-row input[type="date"]`.
- Tests: `actions/tasks_test.go` (alta, edición, inválida, prefill, índice, conservar fecha en toggle) y `models/task_test.go` (opcional, persiste día sin hora).

Propuesta: no rediseñar. Verificar que el código cumple AC1–AC8 y ejecutar las validaciones. Solo se añade algo si una validación revela un fallo.

## Alcance

Incluido:

- Verificación del código ya presente contra los criterios.
- Ejecución de validaciones (gofmt, vet, tests, build).
- Corrección mínima solo si una validación falla por un defecto de esta funcionalidad.

Excluido:

- Hora de vencimiento, recordatorios, ordenación o filtro por fecha.
- Cambios en `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql` y el resto de archivos sin confirmar que no son de este cambio.
- Nuevas dependencias.

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`. Migración ya creada con su `down`.
- **Modelo**: `Task.DueDate *time.Time`. Sin validación extra en `Validate` (la fecha se valida al parsear en la acción).
- **Acción y ruta**: `TasksCreate`, `TasksUpdate`, `TasksEdit`, `TasksIndex`. Rutas existentes, sin cambios.
- **Renderizado**: índice y edición. Sin cambios de ruta.
- **CSS**: estilos mínimos para la fecha y el input. Sin JavaScript nuevo.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Revierte la columna |
| `models/task.go` | Campo `DueDate` |
| `actions/tasks.go` | Parseo, validación y persistencia de la fecha |
| `templates/tasks/index.plush.html` | Campo en alta y fecha mostrada |
| `templates/tasks/edit.plush.html` | Campo en edición |
| `public/assets/app.css` | Estilos de la fecha |
| `actions/tasks_test.go` | Tests de comportamiento de acciones |
| `models/task_test.go` | Tests de persistencia del modelo |

## Implementación paso a paso

1. Confirmar que la rama contiene los archivos de la tabla y que `git diff main...HEAD` solo afecta a esos archivos más los planes.
2. Comprobar AC1–AC8 contra el código y los tests existentes. Anotar cualquier criterio sin test.
3. Si falta un test para un criterio, añadirlo siguiendo el patrón de `actions/tasks_test.go` (`as.authed`, `Post`/`Put`, `as.Contains`). No ampliar el alcance.
4. Ejecutar validaciones (ver siguiente sección). Si algo falla por un defecto de esta feature, corregir lo mínimo y repetir.

## Tests y validación

Condiciones:

- PostgreSQL accesible y migraciones de test aplicadas.
- Tests de aplicación con `GO_ENV=test` (la suite `actions/` usa `todo_test`; sin `GO_ENV` escribe en development).

Comandos, en orden:

1. `gofmt -l .` — no debe listar archivos de este cambio.
2. `go vet ./...` — sin errores.
3. `GO_ENV=test go test -p 1 ./...` — todos los tests pasan; en particular `Test_Tasks*` y `Test_Task_DueDate*`.
4. `go build ./...` — compila.

Tests que demuestran el comportamiento: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`, `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Nota: los tests de alta/edición envían `authenticity_token` mediante `as.authed`; el commit `d222c26` ya lo cubre.

## Riesgos y recuperación

- Si PostgreSQL no está disponible, los tests de acciones y modelo no se pueden ejecutar. Se informa sin abrir PR como verificado.
- Si `GO_ENV` no está definido, los tests escriben en la base de desarrollo. Ejecutar siempre con `GO_ENV=test`.
- Archivos sin confirmar fuera de este cambio (`config/adw.yml`, `.claude/skills/ship/`, `migrations/schema.sql`, `.txt`, `.x`) no se incluyen en el PR.
- Revertir la funcionalidad: ejecutar el `down.fizz` y revertir los commits de la rama.
- Incertidumbre: no se ha comprobado aún si las migraciones de test están aplicadas en este entorno.
