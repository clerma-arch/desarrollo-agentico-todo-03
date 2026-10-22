# Feature: añadir fecha límite (solo día) a las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo ("fechaimie"); la lectura adoptada es **añadir a cada tarea una fecha límite opcional, con precisión de día y sin hora**.

Requisitos explícitos (del cuerpo de la issue):
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

Estado de la rama `feat/issue-3-add-task-due-date` (commits `d13ec2c` y `7913cf0`, aún no en `main`). **La funcionalidad ya está implementada en la rama**; este plan documenta lo implementado y lo que falta verificar. Lo que este plan describía antes como pendiente (struct de entrada solo con `Title`, binding directo de la fecha) ya no aplica.

Implementado:
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `.down.fizz` hace `drop_column`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02")`.
  - `TasksCreate` y `TasksUpdate` leen `due_date` como `string` en su struct de entrada, lo parsean, y si falla añaden el error `dueDateInvalidMessage` a `due_date` y responden 422.
  - `TasksUpdate` asigna `task.DueDate` en cada PUT, así que un PUT sin `due_date` la borra (ver R6).
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el formulario de alta; muestra `Fecha límite: DD/MM/YYYY` solo si `DueDate != nil`.
- `templates/tasks/edit.plush.html`: `<input type="date" name="due_date">` precargado en `YYYY-MM-DD`, con `field-hint` y error.
- `public/assets/app.css`: clase `.task-due` (texto secundario) y `.new-task-row input[type="date"]` (ancho automático).
- Tests:
  - `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
  - `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`.

Propuesta: no hay cambios de diseño. Solo verificar que la rama es correcta y completar cualquier hueco que aparezca en la validación.

## Alcance

Incluido:
- Verificar que la implementación de la rama cumple los criterios 1 a 5.
- Corregir solo lo que falle en la validación (sin ampliar alcance).
- Dejar fuera del PR los cambios no relacionados ya presentes en el árbol de trabajo (`config/adw.yml`, `migrations/schema.sql`, `.claude/skills/ship/`).

Excluido:
- Hora, zona horaria, recordatorios o notificaciones.
- Ordenar o filtrar por fecha; resaltar tareas vencidas.
- Cambios en `TasksToggle`, `TasksDestroy`, rutas, fixtures y seed.
- Cambiar el comportamiento de PUT sin `due_date` (ver R6) salvo que la validación lo exija.

## Diseño por capas

- **Persistencia y migración**: columna `date` nullable, sin backfill.
- **Modelo**: campo `*time.Time`; nil = sin fecha.
- **Acciones**: parseo explícito de la cadena del formulario en `parseDueDate`, reutilizado por alta y edición. Error de formato → 422 con mensaje en `due_date`.
- **Renderizado**: input de tipo `date` en alta y edición; muestra condicional en la lista.
- **CSS**: dos reglas nuevas, reutilizando variables existentes (`--color-muted`).
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` y `.down.fizz`: columna nueva y su reversión.
- `models/task.go`: campo `DueDate`.
- `models/task_test.go`: tests de modelo (persistencia opcional y día sin hora).
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: tests de acciones con `ActionSuite`; `taskForEditing` existente.
- `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`: formularios y lista.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.
- Solo lectura: `.github/workflows/ci.yaml` (orden de validación: `soda migrate -e test`, luego `go test -p 1 ./...`).

## Implementación paso a paso

Ya hecho en la rama. Para completar la entrega:

1. Confirmar que el árbol de la rama contiene los commits `d13ec2c` y `7913cf0` y que `plans/add-task-due-date.md` es el único plan de la feature.
2. Aplicar migraciones de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación (ver siguiente sección). Si algo falla, corregir solo ese punto dentro del alcance.
4. Opcional, si es práctico: `soda migrate down -e test` y `soda migrate -e test` para comprobar la reversión (criterio 5).
5. No tocar `migrations/schema.sql`, `config/adw.yml` ni `.claude/skills/ship/`: no son parte de la feature.

## Tests y validación

Comprobaciones pertinentes, en este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración nueva se aplica sin error.
- `go test -p 1 ./...` → todos pasan, incluidos los 8 tests nuevos de modelo y acciones.
- `go build ./...` → compila.

Requisitos de entorno: PostgreSQL 17 accesible con la configuración de `database.yml` para `test` (en CI: `postgres:17.6-alpine`, usuario y contraseña `postgres`). En esta máquina el puerto 5432 responde, pero no hay `psql` ni `pg_isready`; la conexión de los tests no se ha comprobado todavía. `soda` está instalado (`~/go/bin/soda`) y `go` 1.27.1 también.

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`, `Test_TasksIndex_ListsTasks` (existentes).
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksUpdate_PendingTaskRemainsPending`.
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → `soda migrate -e test` y, si es práctico, la prueba de `down`/`up`.

Estos tests no se han ejecutado en esta sesión de planificación; la validación real es el paso de implementación.

## Riesgos y recuperación

- **R1, binding de fecha**: resuelto en la rama, porque la fecha se lee como cadena y se parsea con `time.Parse`. No depende de la conversión por defecto de Buffalo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve medianoche UTC. Se guarda y se muestra solo año/mes/día, así que no hay desplazamiento de día en la presentación.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas en PostgreSQL. Rollback: `soda migrate down -e test` (o el entorno correspondiente) ejecuta `drop_column`.
- **R4, formato en el navegador**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma; `DD/MM/YYYY` es solo presentación.
- **R5, lectura de la issue**: "fechaimie" es ambiguo. Se ha adoptado fecha límite solo de día. Si la intención era otra, hay que revisar el plan antes de abrir el PR.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` asigna `DueDate` en cada petición. El formulario de edición siempre envía el campo, así que no afecta a la UI, pero un cliente que omita el campo borrará la fecha. Decisión a confirmar: aceptarlo (comportamiento de reemplazo completo del formulario) o conservar la fecha si el campo no llega.
- **R7, archivos ajenos a la feature**: `config/adw.yml` (modificado), `migrations/schema.sql` y `.claude/skills/ship/` (sin seguimiento) existían antes de empezar. No se incluyen en el PR.
- **Recuperación**: si la validación falla, revertir la migración con `soda migrate down -e test`, y corregir o descartar el cambio concreto. Ningún dato existente se modifica porque la columna es nueva y nullable.
