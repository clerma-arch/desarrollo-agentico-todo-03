# Feature: fecha límite de las tareas (solo día)

## Encargo y criterios de aceptación

Issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". Se interpreta "fechaimie" como **fecha límite** (due date) de cada tarea.

Criterios de aceptación:

1. Una tarea puede tener una fecha límite opcional, guardada solo como día (sin hora).
2. La fecha se puede indicar al crear una tarea y al editarla.
3. Dejar el campo vacío significa tarea sin fecha límite.
4. La lista muestra la fecha límite en formato `dd/mm/yyyy` cuando existe.
5. Una fecha inválida no se guarda y se muestra un mensaje de error junto al campo, conservando lo escrito.
6. Completar o reabrir una tarea no modifica su fecha límite.

Requisitos explícitos: fecha opcional, solo día sin hora, visible en la lista y editable.

Inferencias (no estaban en el issue): el formato de error "La fecha no es válida.", la etiqueta "Fecha límite", el texto de ayuda "Solo indica el día, sin hora." y que la fecha no se valida contra "hoy" (se aceptan fechas pasadas).

## Comportamiento actual y propuesta

**Estado de la rama `feat/issue-3-add-task-due-date` (comprobado en el código, no en los planes previos):** la funcionalidad ya está implementada en los commits de la rama, respecto a `main`:

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `.down.fizz`.
- Modelo `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- Acciones `actions/tasks.go`: `parseDueDate` (formato `2006-01-02`, vacío = nil), `TasksCreate` y `TasksUpdate` leen `due_date`, validan con `task.Validate(tx)` y añaden el error `due_date` si el formato no es válido. `TasksEdit` y `TasksIndex` fijan `dueDateInput`.
- Templates: `templates/tasks/index.plush.html` (campo `type="date"` en el formulario de alta, mensaje de error, texto `task-due` en la lista) y `templates/tasks/edit.plush.html` (campo, ayuda y error).
- CSS `public/assets/app.css`: clases `.task-due` y `.new-task-row input[type="date"]`.
- Tests de modelo (`models/task_test.go`) y de acciones (`actions/tasks_test.go`) para crear, editar, validar, mostrar, conservar al completar y limpiar la fecha.

**Cambio pendiente sin confirmar:** la aplicación usa `csrf.New` (`actions/app.go:45`). Las peticiones POST, PUT y DELETE de los tests sin token CSRF devuelven 403. El fichero `.txt` (sin seguimiento) guarda la salida de fallo: `Test_TasksCreate_EmptyTitle_Invalid`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksCreate_OnlySpaces_Invalid`. Los cambios sin confirmar en `actions/actions_test.go` (helpers `authed` y `csrfToken`) y en `actions/tasks_test.go` (sustituyen `as.HTML(...)` por `as.authed(...)` en las peticiones que cambian estado) corrigen esos fallos.

**Propuesta:** no añadir código de aplicación nuevo. Verificar la rama completa, comprometer los tests con el helper CSRF y abrir el PR. Si la verificación encuentra un fallo, corregirlo solo dentro del alcance de esta feature.

## Alcance

Incluido:

- Fecha límite opcional por tarea (día, sin hora) ya presente en la rama.
- Helper CSRF en los tests de acciones (`actions/actions_test.go`, `actions/tasks_test.go`), necesario para que los tests de escritura pasen con `csrf.New` activo.

Excluido:

- `config/adw.yml` (cambio sin confirmar: `repo` y `prompt` del trigger ADW). No es parte de la feature. No se incluye en el PR.
- `.txt` y `migrations/schema.sql` (sin seguimiento; `.txt` es una salida de test antigua y `schema.sql` está vacío). No se incluyen.
- `plans/` de otras pasadas sobre el mismo issue (sin seguimiento). Solo se añade el plan actual, si el flujo lo requiere.
- `.claude/skills/ship/` (sin seguimiento). No se incluye.
- Recordatorios, notificaciones, filtros por fecha, orden por fecha, zona horaria o validación de fechas pasadas.

## Diseño por capas

- **Persistencia:** columna `due_date date null` en `tasks`. Migración ya creada. Se usa `date` (no `timestamp`) porque el requisito es solo día.
- **Modelo:** `models.Task.DueDate *time.Time`, puntero para representar "sin fecha".
- **Acción y ruta:** `TasksCreate` (`POST /tasks`), `TasksUpdate` (`PUT /tasks/{id}`) y `TasksEdit` (`GET /tasks/{id}/edit`). No hay rutas nuevas.
- **Renderizado:** `tasks/index.plush.html` y `tasks/edit.plush.html`. Se muestra `DueDate.Format("02/01/2006")`.
- **CSS:** `.task-due` y el ajuste del campo de fecha en la fila de alta.
- **JavaScript:** no se requiere.

## Archivos relevantes

| Ruta | Responsabilidad | Estado |
|---|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añadir y quitar la columna | Comprometido |
| `models/task.go` | Campo `DueDate` | Comprometido |
| `models/task_test.go` | Tests de fecha opcional y persistencia solo día | Comprometido |
| `actions/tasks.go` | `parseDueDate`, creación y actualización con fecha | Comprometido |
| `actions/tasks_test.go` | Tests de acciones de fecha; uso de `as.authed` | Comprometido + cambio sin confirmar |
| `actions/actions_test.go` | Helpers `authed` y `csrfToken` | Cambio sin confirmar |
| `actions/app.go` | Middleware `csrf.New` (no se modifica) | Referencia |
| `templates/tasks/index.plush.html` | Campo de alta, error y texto de fecha en la lista | Comprometido |
| `templates/tasks/edit.plush.html` | Campo de edición, ayuda y error | Comprometido |
| `public/assets/app.css` | Estilos de fecha | Comprometido |
| `config/adw.yml` | Configuración del trigger ADW | Excluido |
| `.txt`, `migrations/schema.sql` | Artefactos sin seguimiento | Excluidos |

## Implementación paso a paso

1. Confirmar que la rama contiene los commits de la feature frente a `main` (`git diff main --stat`), y que no hay cambios de código de aplicación sin confirmar más allá de `actions/actions_test.go` y `actions/tasks_test.go`.
2. Verificar que `migrations/20261022120000_*` se aplica en la base `todo_test` (`soda migrate -e test`). Si `soda` no está instalado, instalarlo como en `.github/workflows/ci.yaml` (`github.com/gobuffalo/pop/v6/soda@v6.4.1`).
3. Ejecutar la validación de la sección siguiente. Si algún test falla, parar y revisar la causa antes de cambiar nada.
4. Hacer stage explícito solo de: `actions/actions_test.go`, `actions/tasks_test.go`, y, si el flujo lo exige, el plan actual. Los ficheros de la feature ya comprometidos no requieren stage.
5. Commit de los cambios pendientes (helper CSRF) con Conventional Commits, si aplica.
6. Push de `feat/issue-3-add-task-due-date` a `origin` y apertura del PR con `gh pr create` contra `main`.

## Tests y validación

Condiciones: PostgreSQL responde en `127.0.0.1:5432` (comprobado). La base `todo_test` debe existir y tener migraciones aplicadas (`soda migrate -e test`). Usar `GO_ENV=test`.

Comandos (en este orden):

- `gofmt -l .`: debe devolver vacío.
- `go vet ./...`
- `go test -p 1 ./...`: debe pasar, incluidos `Test_TasksCreate_*`, `Test_TasksUpdate_*`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate` y `Test_Task_DueDate_*`.
- `go build ./...`

Comportamiento que demuestran los tests existentes:

- Sin fecha → `DueDate` nil (`models/task_test.go`, `Test_Task_DueDate_Optional`).
- Día sin hora persistido → `2026-10-22` (`Test_Task_DueDate_DayOnly_Persists`).
- Alta con fecha y con fecha inválida (`Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`).
- Edición con fijar y limpiar fecha; fecha inválida no guardada (`Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`).
- Lista muestra la fecha (`Test_TasksIndex_ShowsDueDate`).

No hace falta añadir tests nuevos para la feature: la cobertura existe. Solo se añaden los del helper CSRF si faltara alguno de escritura.

## Riesgos y recuperación

- **Tests de escritura sin token CSRF (403):** es el riesgo principal y está cubierto por el helper sin confirmar. Si el helper no se incluye, los tests fallan. Recuperación: incluir `actions/actions_test.go` y `actions/tasks_test.go` en el stage.
- **Base `todo_test` sin migrar o con migraciones distintas:** los tests fallan al leer `due_date`. Recuperación: `soda migrate -e test`.
- **Zona horaria:** `time.Parse` con formato de fecha produce medianoche UTC. Al leer desde PostgreSQL `date` puede aparecer en otra zona si se formatea con hora local. Los tests usan `UTC` y comparan el formato `2006-01-02`. Riesgo bajo; no se cambia en esta fase.
- **Cambios no relacionados en el working tree:** `config/adw.yml`, `.txt`, `migrations/schema.sql` y `.claude/skills/ship/` no deben entrar en el PR. Recuperación: no hacer stage de ellos.
- **Ambigüedad del título "fechaimie":** la interpretación como fecha límite es una inferencia. Si el encargo era otra cosa, el plan cambia.
- **Sin ejecutar tests en esta fase:** este plan no ejecuta la suite; el estado de verde se confirma en `/implement`.

Siguiente paso: `/implement plans/issue-3-fecha-limite-tareas-feature-2026-10-23.md`
