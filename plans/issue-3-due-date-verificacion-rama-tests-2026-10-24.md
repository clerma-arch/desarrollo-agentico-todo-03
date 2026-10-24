# Feature: fecha límite solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task" (añadir fecha límite a las tareas), cuerpo "vale un dia sin hora".

Requisitos explícitos:

1. Cada tarea puede tener una fecha límite opcional.
2. La fecha es solo día, sin hora.

Criterios de aceptación derivados (inferencias razonables, marcadas como tales):

- [explícito] Al crear una tarea se puede indicar una fecha límite; sin fecha, la tarea se crea igual.
- [explícito] La fecha se guarda y se muestra sin hora.
- [inferido] Al editar una tarea se puede fijar, cambiar o quitar la fecha.
- [inferido] Una fecha con formato inválido se rechaza con error visible y sin persistir la tarea.
- [inferido] Completar o reabrir una tarea no altera su fecha límite.
- [inferido] Los errores de validación conservan la fecha introducida en el formulario.

## Comportamiento actual y propuesta

Estado actual de la rama `feat/issue-3-add-task-due-date` (con commits de código ya presentes sobre `main`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz` añade `due_date` como `date` nullable; el `.down.fizz` la elimina.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` a un día sin hora; `TasksCreate` y `TasksUpdate` leen `due_date` mediante un struct de entrada, validan con `task.Validate(tx)`, añaden el error `due_date` si el parseo falla y persisten con `tx.Create` / `tx.Update`.
- `actions/tasks.go` `TasksEdit`/`TasksIndex` fijan `dueDateInput`, usado por las plantillas para repintar el valor tras un error.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta y etiqueta "Fecha límite: dd/mm/aaaa" en cada tarea.
- `templates/tasks/edit.plush.html`: input `type="date"` con ayuda "Opcional. Solo indica el día, sin hora." y error asociado con `aria-describedby`.
- `public/assets/app.css`: estilos `.task-due` y ancho del input de fecha en `.new-task-row`.
- Tests en `actions/tasks_test.go` y `models/task_test.go` cubren alta con y sin fecha, fecha inválida, conservación de la fecha en errores, edición y limpieza de fecha, toggle, render y prefill. `actions/actions_test.go` añade `authed` (token CSRF) para las peticiones que cambian estado.

Propuesta: no rediseñar nada. El trabajo restante es verificar que la rama cumple los criterios anteriores, ejecutar las validaciones y, si alguna falla, corregir solo lo mínimo necesario dentro de los archivos ya tocados.

## Alcance

Incluido:

- Verificación de los criterios de aceptación contra el código actual de la rama.
- Ejecución de las validaciones: `gofmt`, `go vet`, `go test` de `models/` y `actions/` con `GO_ENV=test`, `go build`.
- Correcciones mínimas si una verificación falla.

Excluido:

- Hora en la fecha límite, recordatorios, ordenación o filtro por fecha, zonas horarias.
- Cambios en `config/adw.yml` (modificación local previa, no forma parte de la funcionalidad).
- Archivos sin seguimiento que no son de código: `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/`, y los planes de `plans/` ya existentes.

## Diseño por capas

- Persistencia: columna `due_date date NULL` en `tasks`, migración ya existente. No se cambia.
- Modelo: campo `DueDate *time.Time`. Validación de título sin cambios; la fecha se valida en la acción porque el parseo ocurre allí.
- Acción: `parseDueDate` en `actions/tasks.go`, reutilizando el patrón de struct de entrada y `verrs.Add` ya usados en el archivo.
- Render: plantillas Plush existentes con el mismo patrón de `aria-describedby` y `field-error` que el título.
- CSS: solo las dos reglas añadidas en `public/assets/app.css`.
- Sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Columna `due_date` |
| `models/task.go` | Campo `DueDate` y validación de título |
| `actions/app.go` | Rutas `POST /tasks`, `GET /tasks/{id}/edit`, `PUT /tasks/{id}`; CSRF habilitado |
| `actions/tasks.go` | Parseo y validación de fecha, alta, edición, errores con valor conservado |
| `actions/tasks_test.go`, `actions/actions_test.go` | Tests de comportamiento y helper CSRF `authed` |
| `models/task_test.go` | Persistencia de fecha nula y de fecha solo día |
| `templates/tasks/index.plush.html` | Input de alta y etiqueta de fecha por tarea |
| `templates/tasks/edit.plush.html` | Input de edición y error de fecha |
| `public/assets/app.css` | `.task-due` y ancho del input de fecha |

## Implementación paso a paso

1. Comprobar `git status --short --branch`. Confirmar que la rama es `feat/issue-3-add-task-due-date` y que los cambios sin confirmar fuera de la funcionalidad son `config/adw.yml` y los archivos sin seguimiento indicados.
2. Revisar cada criterio de aceptación contra el código listado en "Comportamiento actual". Anotar cualquier criterio sin cobertura.
3. Ejecutar las validaciones de la sección siguiente.
4. Si algún criterio no tiene test o el test falla, añadir o corregir el mínimo necesario en los archivos ya listados. No tocar nada fuera de la funcionalidad.
5. Repetir las validaciones y dejar constancia de la salida.

## Tests y validación

Condiciones:

- PostgreSQL accesible con las bases `todo_development` y `todo_test` y las migraciones aplicadas.
- Los tests de `actions/` exigen `GO_ENV=test`. Sin esa variable la app escribe en `todo_development` y la suite lee `todo_test`, lo que produce fallos en cascada que no son un fallo del código (ver memoria `go-tests-need-go-env-test`).

Comandos:

- `gofmt -l .` — en el checkout de Windows lista muchos archivos, incluidos archivos no tocados por la rama. Las versiones committed de los archivos de la funcionalidad salen limpias con `git show HEAD:<ruta> | gofmt -l`. Es un problema de fin de línea CRLF del checkout local, no de la rama. Aplicar la comprobación sobre los archivos de la funcionalidad.
- `go vet ./...` — sin salida, éxito.
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/` — ejecutado durante la verificación: ambos paquetes OK.
- `go build ./...` — éxito.
- `go test -p 1 ./...` completo: incluye `adws/`, que falla en Windows por el `gh` real (ver memoria); fallo ajeno a esta funcionalidad y presente en `main`.

Tests que demuestran el comportamiento (ya presentes): `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksToggle_KeepsDueDate`, `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

## Riesgos y recuperación

- Riesgo: la suite de `actions/` se ejecuta sin `GO_ENV=test` y parece fallar por la funcionalidad. Mitigación: exportar `GO_ENV=test`.
- Riesgo: `gofmt -l` reporta falsos positivos por CRLF. Mitigación: comparar contra `git show HEAD:<ruta>`.
- Riesgo: `TasksCreate` y `TasksUpdate` persisten con `tx.Create`/`tx.Update` en lugar de `ValidateAndCreate`. Es intencionado, porque la validación ocurre antes con `task.Validate(tx)` y el error de fecha se añade a esos mismos errores. Verificar que una petición con título vacío y fecha válida no persiste nada (cubierto por `Test_TasksCreate_EmptyTitle_Invalid` y `Test_TasksCreate_InvalidTitleKeepsDueDate`).
- Recuperación: si una corrección rompe los tests, revertir solo el archivo afectado con `git checkout -- <ruta>` tras revisar su diff; no tocar `config/adw.yml` ni los archivos sin seguimiento.

Siguiente paso: `/implement plans/issue-3-due-date-verificacion-rama-tests-2026-10-24.md`
