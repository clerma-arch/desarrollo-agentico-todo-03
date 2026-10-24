# Feature: fecha límite opcional (solo día) en las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`. Título: "añadir fechaimie alas task". Cuerpo: "vale un dia sin hora". El título es ambiguo; la lectura adoptada es **añadir a cada tarea una fecha límite opcional con precisión de día y sin hora**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día (`YYYY-MM-DD`), sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional: una tarea puede no tener fecha.
- Se indica al crear y al editar.
- Se muestra en la lista y se precarga al editar.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona igual que antes.
2. Crear una tarea con fecha `2026-10-22` la guarda como día, sin hora, y la lista la muestra como `22/10/2026`.
3. Editar permite cambiar o borrar la fecha sin alterar el título ni `Completed`.
4. Una fecha mal formada no se guarda: 422 con mensaje en `due_date`, no un 500.
5. La migración añade una columna nullable; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

Estado real de `feat/issue-3-add-task-due-date` (comprobado en esta planificación):

- La rama está sobre `main` con los commits de la feature (`d13ec2c`, `7913cf0`) más commits de tests y de plans. Tiene el PR #5 abierto (`feat(tasks): añadir fecha límite opcional (solo día) a las tareas`).
- **La funcionalidad ya está implementada en la rama.** Migración, modelo, acciones, plantillas, CSS y tests existen.
- Los tests citados en planes anteriores existen en los archivos: en `actions/tasks_test.go`, `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksToggle_KeepsDueDate`; en `models/task_test.go`, `Test_Task_DueDate_Optional` y `Test_Task_DueDate_DayOnly_Persists`.
- No hay cambios de código sin confirmar relacionados con la feature. El único cambio modificado es `config/adw.yml`, ajeno a la feature.

Lo que esta planificación propone: no rediseñar. Verificar la implementación contra los criterios 1 a 5 con las validaciones de la sección correspondiente y corregir únicamente lo que falle. Si todo pasa, el trabajo pendiente es el de entrega (commit y push del estado validado).

Comportamiento implementado:
- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. Reversión: `drop_column`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"`, `form:"due_date"`, `json:"due_date"`. `nil` significa sin fecha.
- `actions/tasks.go`: `parseDueDate` convierte `""` en `nil` y parsea `2006-01-02`. `TasksCreate` y `TasksUpdate` leen `due_date` como cadena; si el parseo falla, añaden `dueDateInvalidMessage` a `due_date` y responden 422.
- `templates/tasks/index.plush.html`: input `date` en el alta; la lista muestra `Fecha límite: DD/MM/YYYY` solo si hay fecha; error de `due_date` bajo el formulario.
- `templates/tasks/edit.plush.html`: input `date` precargado en `YYYY-MM-DD`, texto de ayuda y error.
- `public/assets/app.css`: `.task-due` (texto secundario con `--color-muted`) y `.new-task-row input[type="date"]`.

## Alcance

Incluido:
- Verificar los criterios 1 a 5 contra el código de la rama.
- Corregir solo fallos de esa verificación.
- Entregar únicamente los archivos de la feature y de este plan.

Excluido:
- Cambios de diseño, de rutas o de validación del título.
- Hora, zona horaria explícita, recurrencia.
- Archivos ajenos al cambio: `config/adw.yml` (modificado), `plans/add-task-due-date.md` (modificado), `migrations/schema.sql` y `.claude/skills/ship/` (sin seguimiento). No se añaden al PR.

## Diseño por capas

- **Persistencia y migraciones**: columna `due_date date NULL` en `tasks`, migración posterior a `20260901000000`.
- **Modelo**: `Task.DueDate *time.Time`.
- **Acción y ruta**: sin rutas nuevas. Las acciones existentes de crear y actualizar parsean la fecha como cadena.
- **Renderizado**: inputs `date` en alta y edición; muestra condicional en la lista.
- **CSS**: dos reglas que reutilizan variables existentes.
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz`: columna y reversión.
- `models/task.go`: campo `DueDate`.
- `models/task_test.go`: tests de modelo.
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: tests de acciones (`ActionSuite`).
- `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`: formularios y lista.
- `public/assets/app.css`: estilos nuevos.
- `database.yml`: base de datos `todo_test` (solo lectura).

## Implementación paso a paso

La implementación existe. Para cerrar la entrega:

1. Confirmar que la rama contiene los commits de la feature sobre `main` y que los archivos de la feature no tienen cambios sin confirmar.
2. Aplicar migraciones de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación de la sección siguiente. Si algo falla, corregir solo ese punto.
4. Comprobar la reversión (criterio 5): `soda migrate down -e test` y luego `soda migrate -e test`.
5. Hacer stage solo de los archivos de la feature y de este plan.

## Tests y validación

En este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración nueva se aplica.
- `go test -p 1 ./...` → todos los tests pasan, incluidos los de fecha límite.
- `go build ./...` → compila.

Requisitos: PostgreSQL accesible con la configuración de `database.yml` para `test` (usuario y contraseña `postgres`, host `127.0.0.1`). No se ha comprobado en la planificación.

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`, `Test_TasksIndex_ListsTasks`.
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksToggle_KeepsDueDate`.
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → `soda migrate -e test` y la comprobación de reversión del paso 4.

## Riesgos y recuperación

- **R1, binding**: la fecha se lee como cadena y se parsea con `time.Parse`; no depende de la conversión por defecto de Buffalo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve medianoche UTC. Se muestra solo día, mes y año, así que no hay desplazamiento de día.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas en PostgreSQL. Reversión con `soda migrate down`.
- **R4, formato**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma del navegador.
- **R5, lectura de la issue**: "fechaimie" es ambiguo. Si la intención era otra, hay que revisar antes de mergear.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` asigna `DueDate` en cada petición. El formulario siempre envía el campo, así que la UI no se ve afectada; un cliente que omita el campo borrará la fecha. Decisión pendiente: aceptar el reemplazo completo o conservar la fecha si el campo no llega.
- **R7, cambios ajenos**: los archivos no relacionados se dejan fuera del PR.
- **Recuperación**: si la validación falla, `soda migrate down -e test` revierte la columna y se corrige o descarta el cambio concreto. Ningún dato existente se modifica.
