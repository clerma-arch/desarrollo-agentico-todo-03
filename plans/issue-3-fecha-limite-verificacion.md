# Feature: verificar la fecha límite (solo día) de las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo; la lectura adoptada es **fecha límite opcional de solo día (`YYYY-MM-DD`) en cada tarea, editable al crear y al editar**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día, sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional.
- Se indica al crear y al editar.
- Se muestra en la lista y se precarga al editar.

Criterios de aceptación:
1. Crear una tarea sin fecha sigue funcionando igual que antes.
2. Crear una tarea con fecha `2026-10-22` la guarda como día y la lista la muestra como `22/10/2026`.
3. Editar permite cambiar o borrar la fecha sin alterar título ni `Completed`.
4. Una fecha mal formada no se guarda: 422 con error en `due_date`, nunca 500.
5. La migración añade una columna nullable; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

La funcionalidad **ya está implementada** en `feat/issue-3-add-task-due-date` (commits `d13ec2c` y `7913cf0` sobre `main`, en `origin`). Este plan no añade funcionalidad: verifica el código contra los criterios, completa huecos de test si los hay y deja explícitas las decisiones pendientes.

Implementado (comprobado en `actions/tasks.go`):
- `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02", raw)`.
- `TasksCreate` (línea ~47) y `TasksUpdate` (línea ~105) leen `due_date` como `string`, lo parsean y, si falla, `verrs.Add("due_date", dueDateInvalidMessage)` → 422.
- `TasksUpdate` asigna `task.DueDate` en cada PUT (R6).
- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: columna `date` nullable.
- `models/task.go`: `DueDate *time.Time`.
- `templates/tasks/index.plush.html` y `edit.plush.html`: `<input type="date" name="due_date">`; lista muestra la fecha solo si existe.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.

Tests declarados (suite `ActionSuite` y de modelo):
- `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
- `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`.

Propuesta: sin cambios de diseño. Ejecutar la validación y corregir solo lo que falle.

## Alcance

Incluido:
- Verificar que la rama cumple los criterios 1 a 5.
- Confirmar que los tests previos al cambio (`Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksUpdate_PendingTaskRemainsPending`) siguen siendo válidos con `due_date`.
- Añadir un test solo si un criterio no tiene cobertura.

Excluido:
- Hora, zona horaria, ordenar o filtrar por fecha, resaltar tareas vencidas.
- R8 (conservar la fecha en el alta tras un 422): mejora de UX fuera de los criterios.
- Cambiar R6 (semántica del PUT sin `due_date`), salvo que un test demuestre un fallo.
- Cambios en `TasksToggle`, `TasksDestroy`, rutas o `fixtures/tasks.toml`.
- Archivos ajenos a la feature: `config/adw.yml`, `plans/add-task-due-date.md`, `migrations/schema.sql`, `.claude/skills/ship/`, y los planes `plans/issue-3-verificar-fecha-limite.md` y `plans/verificar-fecha-limite-tareas.md`. No entran en el PR.

## Diseño por capas

Capas afectadas (ya implementadas; solo se verifican):
- **Persistencia**: columna `due_date date NULL`, sin backfill.
- **Modelo**: `*time.Time`; `nil` = sin fecha.
- **Acción**: `parseDueDate` compartido por alta y edición; error de formato → 422.
- **Renderizado**: input `date` en alta y edición; texto condicional en la lista.
- **CSS**: dos reglas de estilo. Sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añadir y revertir la columna |
| `models/task.go` | Campo `DueDate` |
| `models/task_test.go` | Tests de modelo |
| `actions/tasks.go` | `parseDueDate`, `TasksCreate`, `TasksUpdate` |
| `actions/tasks_test.go` | Tests de acción y render |
| `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html` | Formularios y lista |
| `public/assets/app.css` | Estilo de fecha |

## Implementación paso a paso

1. Confirmar que la rama contiene `d13ec2c` y `7913cf0` y que los archivos de la tabla no tienen cambios sin confirmar.
2. Preparar la base de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación de la sección siguiente.
4. Comprobar la reversión (criterio 5): `soda migrate down -e test` y luego `soda migrate -e test`.
5. Si un criterio no tiene test, añadirlo siguiendo el patrón de `actions/tasks_test.go`. No añadir tests fuera de los criterios.
6. Hacer stage solo de los archivos de la feature y de este plan.

## Tests y validación

Comandos, en el orden del CI:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → sin errores.
- `go test -p 1 ./...` → todo en verde, incluidos los 8 tests de fecha límite.
- `go build ./...` → compila.
- Reversión: `soda migrate down -e test` seguido de `soda migrate -e test`.

Condiciones necesarias:
- Go instalado (versión de `go.mod`).
- PostgreSQL en `127.0.0.1:5432` con usuario y contraseña `postgres` y base `todo_test` (`database.yml`). En CI se usa `postgres:17.6-alpine`. Levantarlo con `docker compose up -d db` si no está disponible.
- `soda` instalado (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`).

Cobertura por criterio:
- Criterio 1 → `Test_TasksCreate_Valid`, `Test_Task_DueDate_Optional`.
- Criterio 2 → `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`.
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → paso 2 y reversión del paso 4.

Pendiente de confirmar al ejecutar: que `Test_TasksUpdate_PersistsTitleOnly` no asume que un PUT sin `due_date` conserve la fecha. Si lo asume, R6 y el test se contradicen y hay que decidir cuál cambia.

## Riesgos y recuperación

- **R6, PUT sin `due_date` borra la fecha.** El formulario de edición siempre envía el campo, así que la UI no se ve afectada; un cliente externo que omita el campo pierde la fecha. Decisión por defecto: mantener el reemplazo completo. Recuperación: condicionar la asignación a que la clave llegue.
- **R8, fecha vacía tras 422 en el alta.** `index.plush.html` no reenvía `task.DueDate` a `value`. Solo UX. Recuperación: renderizar `value` como en la edición.
- **R2, zona horaria.** `time.Parse` sin zona devuelve medianoche UTC; al guardar en columna `date` no hay desplazamiento. Revisar si cambia la zona horaria de la base.
- **Migración.** `ADD COLUMN ... NULL` no reescribe la tabla en PostgreSQL. Recuperación: `soda migrate down -e <entorno>`.
- **Entorno sin PostgreSQL.** Si la base no se levanta, los tests no se pueden ejecutar; reportarlo sin declarar la validación superada.
- **R5, lectura de la issue.** "fechaimie" es ambiguo. Si la intención era otra, revisar antes de abrir el PR.
