# Feature: Fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

**Encargo explícito** (issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"; leído con `gh issue view 3`, estado OPEN, etiqueta `adw`):

- Las tareas pueden tener una fecha límite opcional.
- La fecha es solo de día: no tiene hora.

**Criterios de aceptación** (verificables por tests o inspección):

1. Crear una tarea sin fecha funciona como antes y guarda `due_date` vacío.
2. Crear una tarea con fecha `YYYY-MM-DD` guarda ese día sin hora.
3. Una fecha con formato inválido no crea la tarea y muestra `La fecha no es válida.` en el campo `due_date`.
4. Un título inválido al crear o editar conserva la fecha introducida en el formulario.
5. La lista muestra `Fecha límite: DD/MM/AAAA` solo cuando la tarea tiene fecha.
6. El formulario de edición precarga la fecha existente en formato `YYYY-MM-DD`.
7. Editar una tarea puede establecer, cambiar o borrar la fecha; borrar el campo deja la tarea sin fecha.
8. La migración añade la columna `due_date` de tipo `date`, anulable, y tiene su `down`.

**Inferencias** (no explícitas en la issue, tomadas del código ya escrito):

- El campo es un `<input type="date">` opcional. No se valida que la fecha sea futura ni se compara con "hoy".
- No hay indicador de vencida, ordenación ni filtros por fecha: quedan fuera.

## Comportamiento actual y propuesta

**Estado verificado de la rama `feat/issue-3-add-task-due-date`** (leído el código en esta sesión; la rama está sincronizada con `origin/feat/issue-3-add-task-due-date` y `git status` solo muestra cambios ajenos: `config/adw.yml` modificado y `.claude/skills/ship/` sin seguimiento):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: `add_column("tasks", "due_date", "date", {"null": true})` y su `drop_column` inverso.
- Modelo `models/task.go`: `DueDate *time.Time` con `json:"due_date" db:"due_date" form:"due_date"`.
- Acciones `actions/tasks.go`:
  - `parseDueDate` convierte `YYYY-MM-DD` en `*time.Time` (cadena vacía → `nil`).
  - `TasksCreate` lee `title` y `due_date` en una struct de entrada, valida con `task.Validate(tx)`, añade el error `due_date` si el parseo falla y persiste con `tx.Create` solo si no hay errores.
  - `TasksUpdate` hace lo mismo con `tx.Update`.
- Plantillas:
  - `templates/tasks/index.plush.html`: input `type="date"` en el alta y `Fecha límite: DD/MM/AAAA` bajo el título.
  - `templates/tasks/edit.plush.html`: campo de fecha con ayuda y error.
- Estilos `public/assets/app.css`: clase `.task-due` y ancho automático del input de fecha en `.new-task-row`.
- Tests: `actions/tasks_test.go` (+109 líneas) y `models/task_test.go` (+32 líneas).

**Propuesta**: no rediseñar. Lo que falta es de cierre:

1. Confirmar, leyendo los tests, que cada criterio de aceptación tiene al menos un test o una comprobación.
2. Añadir solo el test que falte, si lo hay.
3. Ejecutar la validación.
4. Hacer stage explícito de los archivos del cambio y abrir el PR.

## Alcance

**Incluido**

- Verificar contra los criterios de aceptación el código ya presente en la rama.
- Añadir un test solo si un criterio no tiene ninguno que lo demuestre.
- Commit y PR únicamente con los archivos del cambio (ver Implementación).

**Excluido**

- Indicadores de fecha vencida, ordenación o filtrado por fecha.
- Validar que la fecha sea futura o compararla con hoy.
- Zona horaria del usuario: la fecha se guarda como día calendario sin hora.
- `migrations/schema.sql` (sin seguimiento, no relacionado con la migración de la issue).
- `config/adw.yml`, `.claude/skills/ship/` y los planes antiguos `plans/issue-3-*` y `plans/verificar-fecha-limite-tareas.md` y `plans/add-task-due-date.md`: no forman parte de este cambio.

## Diseño por capas

- **Persistencia y migraciones**: columna `due_date date NULL` en `tasks`. Ya implementado.
- **Modelo**: `Task.DueDate *time.Time`; `nil` representa "sin fecha". Ya implementado.
- **Acción y ruta**: `TasksCreate` y `TasksUpdate` reutilizan `parseDueDate` y el patrón existente de `task.Validate(tx)` + `verrs.Add`. Ya implementado.
- **Renderizado**: `index` y `edit` muestran y precargan la fecha con `Format`. Ya implementado.
- **CSS**: `.task-due` y ajuste del input de fecha en la fila de alta. Ya implementado.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade la columna `due_date`. |
| `migrations/20261022120000_add_due_date_to_tasks.down.fizz` | Elimina la columna. |
| `models/task.go` | Campo `DueDate` del modelo. |
| `models/task_test.go` | Tests de la validación/modelo con fecha. |
| `actions/tasks.go` | `parseDueDate`, lectura y persistencia de la fecha en alta y edición. |
| `actions/tasks_test.go` | Tests de alta y edición con fecha. |
| `templates/tasks/index.plush.html` | Input de fecha en el alta y fecha mostrada en la lista. |
| `templates/tasks/edit.plush.html` | Campo de fecha, ayuda y error en la edición. |
| `public/assets/app.css` | Estilos `.task-due` y del input de fecha. |

## Implementación paso a paso

1. Comprobar que la rama es `feat/issue-3-add-task-due-date` y que está al día con `origin`.
2. Leer `actions/tasks_test.go` y `models/task_test.go` y hacer una tabla criterio → test. Para cada criterio 1–8, anotar el test que lo cubre o marcarlo como hueco.
3. Si hay huecos, añadir solo los tests necesarios siguiendo el estilo de los tests existentes de `actions/tasks_test.go`. No tocar código de aplicación salvo que un test demuestre un fallo real; en ese caso, detenerse e informar.
4. Ejecutar la validación (ver Tests y validación) y registrar la salida.
5. Hacer stage explícito solo de los archivos del cambio:
   - `migrations/20261022120000_add_due_date_to_tasks.up.fizz`
   - `migrations/20261022120000_add_due_date_to_tasks.down.fizz`
   - `models/task.go`, `models/task_test.go`
   - `actions/tasks.go`, `actions/tasks_test.go`
   - `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`
   - `public/assets/app.css`
   - este plan: `plans/issue-3-fecha-limite-solo-dia-cierre-ship.md`
   Nunca `git add .` ni `git add -A`.
6. Si hay tests nuevos, un commit `test(tasks): …` con referencia a la issue. Si no, no crear commit nuevo: el PR parte de los commits ya existentes.
7. Push a `origin` sin `--force`.
8. Abrir el PR con `gh pr create --base main` con título Conventional Commits y descripción con resumen, ruta del plan, tests añadidos y resultado de la validación.

## Tests y validación

Condiciones: PostgreSQL disponible y la base de test con migraciones aplicadas (`GO_ENV=test`).

Comprobaciones pertinentes:

- `gofmt -l .`: no debe listar archivos modificados por el cambio.
- `go vet ./...`
- `go build ./...`
- `GO_ENV=test go test -count=1 -p 1 ./actions/ ./models/`: los paquetes afectados.
- `go test -p 1 ./...`: en la sesión de planificación del plan previo, `./adws/` falló por motivos ajenos al cambio. Registrar la salida tal cual; no es motivo para ocultarla, y si falla, indicar que es previo y no relacionado.

Comportamiento que deben demostrar los tests (criterios 1–8): alta sin fecha, alta con fecha, fecha inválida, título inválido con fecha conservada, precarga en edición, cambio y borrado de fecha, y migración con `down`.

## Riesgos y recuperación

- **Estado no verificado de tests**: esta sesión no ejecutó la suite; la validación se hace en la fase de implementación. Si algún test de `actions` o `models` falla, no abrir el PR.
- **Huecos de cobertura**: si un criterio no tiene test, añadirlo; no ampliar el alcance.
- **Cambios ajenos**: `config/adw.yml` modificado y los planes sin seguimiento no entran en el PR. Avisar en el reporte final.
- **Recuperación**: si el push o el PR fallan, no reescribir historia; corregir y volver a intentar sin `--force`.
- **Migración**: `due_date` es anulable, así que la migración es segura sobre datos existentes. El `down` elimina la columna y pierde las fechas guardadas; es aceptable solo para revertir.
