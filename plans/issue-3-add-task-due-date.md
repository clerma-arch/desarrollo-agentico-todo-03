# Feature: añadir fecha límite (solo día) a las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo ("fechaimie"); la lectura adoptada es **añadir a cada tarea una fecha límite opcional, con precisión de día y sin hora**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día (`YYYY-MM-DD`), sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional: una tarea puede no tener fecha.
- Se puede indicar la fecha al crear y al editar una tarea.
- La fecha se muestra en la lista y se precarga al editar.

Criterios de aceptación:
1. Crear una tarea sin fecha sigue funcionando igual que antes.
2. Crear una tarea con fecha `2026-10-22` la guarda como día, sin hora, y la lista la muestra como `22/10/2026`.
3. Editar una tarea permite cambiar o borrar la fecha sin alterar título ni estado `Completed`.
4. Una fecha mal formada no se guarda: la respuesta es 422 con mensaje en `due_date`, no un 500.
5. La migración añade la columna nullable; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` respecto a `main` (commits `d13ec2c` y `7913cf0`, aún no mergeados): **la funcionalidad ya está implementada en la rama**. Este plan documenta esa implementación, la verifica contra los criterios y define qué queda por validar. No hay cambios de diseño.

Implementado en la rama:
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `.down.fizz` hace `drop_column("tasks", "due_date")`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"`, `form:"due_date"` y `json:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02", raw)`.
  - `TasksCreate` y `TasksUpdate` leen `due_date` como `string` en su struct de entrada, lo parsean y, si falla, añaden `dueDateInvalidMessage` a `due_date` y responden 422.
  - `TasksCreate` crea la tarea con `task.Validate(tx)` + `tx.Create`; `TasksUpdate` asigna `task.DueDate` en cada PUT (ver R6).
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el formulario de alta; la lista muestra `Fecha límite: DD/MM/YYYY` solo si `DueDate != nil`; el error de `due_date` se muestra bajo el formulario.
- `templates/tasks/edit.plush.html`: `<input type="date" name="due_date">` precargado en `YYYY-MM-DD`, con texto de ayuda y error.
- `public/assets/app.css`: clase `.task-due` (texto secundario con `--color-muted`) y `.new-task-row input[type="date"]` (ancho automático).
- Tests añadidos: `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`) y `actions/tasks_test.go` (`Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`).

Propuesta: verificar que la implementación cumple los criterios 1 a 5 mediante las validaciones de la sección siguiente, y corregir solo lo que falle.

## Alcance

Incluido:
- Verificar que la implementación de la rama cumple los criterios 1 a 5.
- Corregir solo lo que falle en la validación, sin ampliar el alcance.
- Incluir en el PR los archivos de la feature y este plan.

Excluido:
- Cualquier cambio de diseño, de rutas o de validación de título.
- Los cambios de árbol de trabajo no relacionados: `config/adw.yml` (modificado), `plans/add-task-due-date.md` (modificado, plan previo del usuario), `migrations/schema.sql` y `.claude/skills/ship/` (sin seguimiento). No se incluyen en el PR.
- Hora, zona horaria explícita y recurrencia de fechas.

## Diseño por capas

- **Persistencia y migraciones**: columna `due_date date NULL` en `tasks`. Migración nueva con timestamp posterior a la existente (`20261022120000`).
- **Modelo**: `Task.DueDate *time.Time`; nil = sin fecha.
- **Acción y ruta**: sin rutas nuevas. `TasksCreate` y `TasksUpdate` parsean la fecha como cadena y devuelven 422 con error en `due_date`.
- **Renderizado**: input `date` en alta y edición; muestra condicional en la lista.
- **CSS**: dos reglas nuevas que reutilizan variables existentes (`--color-muted`).
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` y `.down.fizz`: columna nueva y su reversión.
- `models/task.go`: campo `DueDate`.
- `models/task_test.go`: tests de modelo (persistencia opcional y día sin hora).
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: tests de acciones con `ActionSuite`.
- `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`: formularios y lista.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.
- `database.yml`: configuración de la base de datos `todo_test` (solo lectura).
- `.github/workflows/ci.yaml`: orden de validación de CI (solo lectura).

## Implementación paso a paso

La implementación ya existe en la rama. Pasos para completar la entrega:

1. Confirmar que la rama contiene `d13ec2c` y `7913cf0` sobre `main` y que el árbol de la feature no tiene cambios sin confirmar.
2. Aplicar migraciones de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación de la sección siguiente. Si algo falla, corregir solo ese punto dentro del alcance.
4. Opcional: `soda migrate down -e test` y después `soda migrate -e test` para comprobar la reversión (criterio 5).
5. Hacer stage solo de los archivos de la feature y de este plan.

## Tests y validación

Comprobaciones, en este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración nueva se aplica sin error.
- `go test -p 1 ./...` → todos los tests pasan, incluidos los nuevos de modelo y acciones.
- `go build ./...` → compila.

Requisitos de entorno: PostgreSQL accesible con la configuración de `database.yml` para `test` (usuario y contraseña `postgres`, host `127.0.0.1` por defecto). `go` 1.27.1 está instalado. La conexión a la base de datos de test no se ha comprobado en la planificación; la ejecución real es el paso de validación.

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`, `Test_TasksIndex_ListsTasks` (existentes).
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksUpdate_PendingTaskRemainsPending`.
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → `soda migrate -e test` y, si es práctico, `down`/`up`.

## Riesgos y recuperación

- **R1, binding de fecha**: la fecha se lee como cadena y se parsea con `time.Parse`; no depende de la conversión por defecto de Buffalo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve medianoche UTC. Se guarda y se muestra solo año, mes y día, así que no hay desplazamiento de día en la presentación.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas en PostgreSQL. Reversión: `soda migrate down -e <entorno>` ejecuta `drop_column`.
- **R4, formato en el navegador**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma; `DD/MM/YYYY` es solo presentación.
- **R5, lectura de la issue**: "fechaimie" es ambiguo. Se adopta fecha límite solo de día. Si la intención era otra, hay que revisar el plan antes de mergear.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` asigna `DueDate` en cada petición. El formulario de edición siempre envía el campo, así que la UI no se ve afectada; un cliente que omita el campo borrará la fecha. Decisión pendiente: aceptar el reemplazo completo o conservar la fecha si el campo no llega.
- **R7, cambios ajenos**: `config/adw.yml`, `plans/add-task-due-date.md`, `migrations/schema.sql` y `.claude/skills/ship/` existían antes de empezar y no se incluyen en el PR.
- **Recuperación**: si la validación falla, revertir la migración con `soda migrate down -e test` y corregir o descartar el cambio concreto. Ningún dato existente se modifica porque la columna es nueva y nullable.
