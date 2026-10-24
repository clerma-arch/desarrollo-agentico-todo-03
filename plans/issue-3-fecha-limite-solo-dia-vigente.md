# Feature: fecha límite de las tareas (solo día) — issue #3

Plan vigente para la issue #3 (`añadir fechaimie alas task`, cuerpo: «vale un dia sin hora»). Sustituye a los planes previos de `plans/issue-3-*.md`, que quedan como historial y no deben ejecutarse de nuevo.

## Encargo y criterios de aceptación

Encargo explícito de la issue:

- Las tareas pueden tener una fecha límite.
- La fecha límite es solo un día, sin hora.

Criterios de aceptación derivados (comprobables):

1. Crear una tarea con fecha: se guarda el día y se muestra como `Fecha límite: DD/MM/AAAA` en la lista.
2. Crear una tarea sin fecha: se guarda sin fecha y no se muestra la línea de fecha límite.
3. Editar una tarea: se puede fijar, cambiar y borrar la fecha; el campo aparece prerrellenado con el día actual.
4. Una fecha mal formada no se guarda y muestra el error `La fecha no es válida.` en el formulario correspondiente, conservando el título introducido.
5. Un título inválido al crear o editar no pierde la fecha introducida.
6. Completar, reabrir o alternar una tarea conserva su fecha límite.
7. Migración: la columna `due_date` es de tipo `date` y anulable; las tareas existentes quedan sin fecha.

Inferencias (no explícitas en la issue, asumidas):

- El formato de entrada es el del input nativo `type="date"` (`YYYY-MM-DD`) y el de visualización en lista es `DD/MM/AAAA`.
- No hay fecha por defecto ni validación de fechas pasadas; la issue no lo pide.

## Comportamiento actual y propuesta

Estado verificado en la rama `feat/issue-3-add-task-due-date` (commits `69e8826`…`5332a58` y anteriores), frente a `main` (`935a5e6`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: añade y elimina `due_date` (`date`, anulable). Hecha.
- `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. Hecho.
- `actions/tasks.go`: `parseDueDate` (vacío → `nil`; formato inválido → error), `TasksCreate` y `TasksUpdate` leen `due_date`, validan con `task.Validate(tx)` y añaden `due_date` a `verrs` si el parseo falla. Persistencia con `tx.Create` / `tx.Update`. Hecho.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de alta, línea `Fecha límite` en cada tarea, error `due_date`. Hecho.
- `templates/tasks/edit.plush.html`: campo de fecha con ayuda y error asociado, prerrellenado. Hecho.
- `public/assets/app.css`: clases `.task-due` y ajuste de `input[type="date"]` en `.new-task-row`. Hecho.
- Tests de modelo (`models/task_test.go`) y de acciones (`actions/tasks_test.go`): cubren alta y edición con/sin fecha, fecha inválida, título inválido conservando fecha, alternar conservando fecha y prerrelleno. Hecho.

No hay cambios de código pendientes identificados frente a los criterios. Lo que falta es verificación y entrega, no implementación.

Cambio propuesto: ninguno en código. Ejecutar las validaciones de la sección siguiente y, si todo pasa, preparar el PR.

## Alcance

Incluido:

- Verificar los criterios 1 a 7 con los tests existentes y con la validación manual de la pantalla.
- Entregar la rama con el plan vigente.

Excluido:

- Cambios de diseño o de comportamiento no pedidos (recordatorios, fechas pasadas, orden por fecha, hora).
- Modificar `config/adw.yml` (cambio local no relacionado, ver Riesgos).
- Incluir planes antiguos o `migrations/schema.sql` en el PR.

## Diseño por capas

- **Persistencia y migraciones**: ya aplicada. Sin cambios.
- **Modelo**: ya aplicado. `Validate` no valida la fecha porque el parseo se hace en la acción; es el patrón que se mantiene.
- **Acción y ruta**: ya aplicadas. Se mantiene el patrón de `Bind` en struct local, igual que `TasksUpdate` previo.
- **Renderizado**: ya aplicado en las dos plantillas.
- **CSS**: ya aplicado; sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Columna `due_date` |
| `models/task.go` | Campo `DueDate` |
| `models/task_test.go` | Tests de modelo |
| `actions/tasks.go` | `parseDueDate`, alta y edición |
| `actions/tasks_test.go` | Tests de alta, edición, alternar y errores |
| `templates/tasks/index.plush.html` | Alta y lista |
| `templates/tasks/edit.plush.html` | Edición |
| `public/assets/app.css` | Estilos de fecha |
| `config/adw.yml` | **No pertenece al PR** (cambio local preexistente) |

## Implementación paso a paso

1. Confirmar que la rama contiene exactamente los archivos de la tabla anterior: `git diff --stat main...HEAD`.
2. Confirmar que no hay cambios de código sin confirmar: `git status --short` solo debe mostrar `config/adw.yml`, `.claude/skills/ship/`, `.txt`, `migrations/schema.sql` y los planes sin seguimiento (todos fuera del PR, salvo este plan).
3. Preparar la base de test como CI: migraciones de `todo_test` aplicadas (`soda migrate -e test` con `GO_ENV=test`).
4. Ejecutar las validaciones de la sección siguiente. Si alguna falla, detenerse (no es un cambio de alcance; ver Riesgos).
5. Verificar en la app los criterios 1 a 6 (ver `/run` o `env:start`).
6. Stage explícito del plan y de los archivos de la tabla; commit y push si hace falta.

## Tests y validación

Requisitos previos:

- PostgreSQL en `127.0.0.1:5432` con `todo_test`.
- `GO_ENV=test` en todas las ejecuciones de `go test`. Sin esa variable, las pruebas de acciones devuelven 403 por el middleware CSRF (`actions/app.go`), como en `main`. CI también la define.
- La base de test debe estar limpia o las pruebas de lista fallan: en la ejecución de este plan `Test_TasksIndex_EmptyState` falló porque `todo_test` contenía 71 tareas residuales. Limpiar o recrear la base antes de ejecutar.

Comandos (ya verificados en esta sesión con `GO_ENV=test`):

- `go build ./...`: OK (exit 0).
- `go vet ./actions/... ./models/... ./cmd/...`: sin salida (OK).
- `go test -count=1 -p 1 ./actions/... ./models/...`: `ok todo/actions`, `ok todo/models`.
- `gofmt -l`: **avisa sobre casi todos los archivos Go**, incluidos los no tocados por la rama. Causa: el working tree tiene finales CRLF (`git ls-files --eol` muestra `i/lf w/crlf`), mientras que el índice es LF. El contenido de `main` pasa `gofmt -l` sin aviso. No es un error de formato del cambio; ejecutar `gofmt -l` sobre la copia LF o con `core.autocrlf` adecuado antes de juzgarlo.

Comprobaciones pendientes para dar el cambio por cerrado:

- Pruebas de fecha: 32 coincidencias de `DueDate|due` en la salida verbose de `actions` y `models`.
- Verificación manual de criterios 1 a 6 en la app.

## Riesgos y recuperación

- **Working tree sucio**: `config/adw.yml` tiene un cambio local (`repo` y `prompt` del trigger) que no pertenece a esta issue. Dejarlo fuera del PR y avisar al usuario.
- **Planes sin seguimiento**: hay decenas de `plans/issue-3-*.md` sin commitear y otros ya commiteados. Solo se incluye este plan; los demás quedan fuera.
- **CRLF**: el aviso de `gofmt` puede confundir la revisión. Si el PR se revisa en Windows, usar `git diff --ignore-cr-at-eol` o normalizar finales de línea antes de abrir el PR. No se normaliza en esta fase para no tocar archivos fuera de alcance.
- **Base de test contaminada**: los fallos de lista se deben a datos residuales, no al código. Recuperación: recrear `todo_test` y repetir las migraciones.
- **Rollback**: `migrations/20261022120000_add_due_date_to_tasks.down.fizz` elimina la columna.
