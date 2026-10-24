# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo ("fechaimie"); la lectura adoptada es **añadir a cada tarea una fecha límite opcional, con precisión de día y sin hora**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día (`YYYY-MM-DD`), sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional: una tarea puede no tener fecha.
- Se puede indicar la fecha al crear y al editar una tarea.
- La fecha se muestra en la lista y se precarga en el formulario de edición.

Criterios de aceptación:
1. Crear una tarea sin fecha sigue funcionando igual que antes.
2. Crear una tarea con fecha `2026-10-22` la guarda como día, sin hora, y la lista la muestra como `22/10/2026`.
3. Editar una tarea permite cambiar o borrar la fecha sin alterar el título ni el estado `Completed`.
4. Una fecha mal formada no se guarda: la respuesta es 422 con mensaje en `due_date`, no un 500.
5. La migración añade la columna nullable; las tareas existentes quedan sin fecha y la migración es reversible.

## Comportamiento actual y propuesta

Estado verificado de la rama `feat/issue-3-add-task-due-date` frente a `main`: la rama tiene 8 commits por delante de `main`, entre ellos `d13ec2c feat(tasks): añadir fecha límite opcional (solo día) a las tareas` y `69e8826 fix(tasks): conservar la fecha límite al fallar la validación del alta`. **La funcionalidad ya está implementada y confirmada en el código committeado de la rama.** Este plan no propone cambios de diseño: documenta el estado, define la validación final y limita cualquier corrección a lo que falle.

Implementado en la rama (comprobado leyendo el código):
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `.down.fizz` hace `drop_column("tasks", "due_date")`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"`, `form:"due_date"` y `json:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02", raw)`.
  - `TasksCreate` y `TasksUpdate` leen `due_date` como `string` en su struct de entrada, lo parsean y, si falla, añaden `dueDateInvalidMessage` a `due_date` y responden 422 re-renderizando el formulario.
  - `TasksUpdate` asigna `task.DueDate` en cada petición (ver R6).
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el formulario de alta; la lista muestra `Fecha límite: DD/MM/YYYY` solo si `DueDate != nil`; el error de `due_date` aparece bajo el formulario.
- `templates/tasks/edit.plush.html`: `<input type="date" name="due_date">` precargado en `YYYY-MM-DD`, con texto de ayuda y error.
- `public/assets/app.css`: clase `.task-due` (texto secundario con `--color-muted`) y `.new-task-row input[type="date"]` (ancho automático).
- Tests committeados: `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`) y `actions/tasks_test.go` (`Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`).

Cambios sin confirmar ajenos a la feature, que **no forman parte del PR**:
- `config/adw.yml` (modificado), `migrations/schema.sql`, `.claude/skills/ship/` y los planes de `plans/` sin seguimiento distintos de este.
- En la verificación, `actions/tasks_test.go` no tiene cambios sin confirmar: `Test_TasksToggle_KeepsDueDate` ya está committeado.

Propuesta: verificar que el código committeado cumple los criterios 1 a 5 con las validaciones de la sección siguiente, y corregir solo lo que falle.

## Alcance

Incluido:
- Verificar los criterios 1 a 5 contra el código committeado de la rama.
- Corregir únicamente un fallo detectado en esa verificación, sin ampliar el alcance.
- Abrir el PR con los commits ya existentes de la rama más este plan.

Excluido:
- Cambios de diseño, de rutas o de validación del título.
- Hora, zona horaria explícita, recurrencia de fechas y filtros por fecha.
- Cualquier archivo sin seguimiento o modificado ajeno a la feature (ver lista anterior).
- Merge y push forzado.

## Diseño por capas

- **Persistencia y migraciones**: columna `due_date date NULL` en `tasks`, en una migración con timestamp posterior a la existente (`20261022120000`).
- **Modelo**: `Task.DueDate *time.Time`; `nil` significa sin fecha.
- **Acción y ruta**: sin rutas nuevas. `TasksCreate` y `TasksUpdate` parsean la fecha como cadena y devuelven 422 con el error en `due_date`.
- **Renderizado**: input `date` en el alta y en la edición; la lista muestra la fecha solo si existe.
- **CSS**: dos reglas que reutilizan `--color-muted`.
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` y `.down.fizz`: columna nueva y su reversión.
- `models/task.go`: campo `DueDate`.
- `models/task_test.go`: tests de modelo (opcional y día sin hora).
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: tests de acciones con `ActionSuite` (versión committeada; no tocar la versión sin seguimiento).
- `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`: formularios y lista.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.
- `database.yml`: base `todo_test` con usuario `postgres` y host `127.0.0.1` por defecto (solo lectura).

## Implementación paso a paso

La implementación ya está en la rama. Pasos para cerrar la entrega:

1. Confirmar que `git log main..HEAD` contiene los commits de la feature y que `git status` solo muestra los cambios ajenos listados arriba.
2. Aplicar migraciones de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación de la sección siguiente. Si un criterio falla, corregir solo ese punto dentro del alcance y repetir la validación.
4. Comprobar la reversión (criterio 5): `soda migrate down -e test` y después `soda migrate -e test`.
5. Hacer stage solo de los archivos de la feature y de este plan, con una lista explícita, y hacer push a `origin` sin forzar.

## Tests y validación

Comprobaciones, en este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración nueva se aplica sin error.
- `go test -p 1 ./...` → todos los tests pasan, incluidos los de modelo y acciones de la feature.
- `go build ./...` → compila.
- Reversión de la migración (paso 4 anterior).

Requisitos de entorno: PostgreSQL accesible con la configuración de `database.yml` para `test`. `go` está instalado en `C:\Program Files\Go\bin`. No se ha comprobado todavía que `soda` ni PostgreSQL estén disponibles; si no lo están, la validación de base de datos no puede ejecutarse y se debe informar como no verificada.

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`, `Test_TasksIndex_ListsTasks` (existentes).
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksUpdate_PendingTaskRemainsPending`. (El test de toggle que conserva la fecha, `Test_TasksToggle_KeepsDueDate`, está committeado en `actions/tasks_test.go`.)
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → paso 4 (`soda migrate down` y `up`).

## Riesgos y recuperación

- **R1, binding de fecha**: la fecha se lee como cadena y se parsea con `time.Parse`; no depende de la conversión por defecto de Buffalo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve medianoche UTC. La presentación usa solo día, mes y año, así que no hay desplazamiento de día.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas en PostgreSQL. Reversión: `soda migrate down -e <entorno>`.
- **R4, formato en el navegador**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma; `DD/MM/YYYY` es solo presentación.
- **R5, lectura de la issue**: "fechaimie" es ambiguo. Se adopta fecha límite solo de día. Si la intención era otra, hay que revisar el plan antes de mergear.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` asigna `DueDate` en cada petición. El formulario de edición siempre envía el campo, así que la UI no se ve afectada; un cliente que omita el campo borra la fecha. Decisión pendiente del usuario: aceptar el reemplazo completo (comportamiento actual) o conservar la fecha si el campo no llega. Este plan no la cambia.
- **R7, cambios ajenos sin confirmar**: `config/adw.yml` tiene cambios sin confirmar y hay planes sin seguimiento ajenos a la feature; no se incluyen en el PR. Si se quieren en el PR, deben confirmarse en una decisión aparte.
- **Recuperación**: si la validación falla, revertir la migración con `soda migrate down -e test` y corregir solo el punto concreto. Ningún dato existente se modifica porque la columna es nueva y nullable.
