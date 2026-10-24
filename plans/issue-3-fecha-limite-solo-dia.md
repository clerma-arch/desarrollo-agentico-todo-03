# Feature: fecha límite opcional (solo día) en las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo; la lectura adoptada es **cada tarea puede tener una fecha límite opcional, con precisión de día (`YYYY-MM-DD`) y sin hora**.

Requisitos explícitos (del cuerpo de la issue):
- La fecha es solo de día, sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional: una tarea puede no tener fecha.
- Se indica al crear y al editar.
- La lista la muestra como `DD/MM/YYYY` y la edición la precarga.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona como antes.
2. Crear una tarea con `2026-10-22` la guarda como día sin hora y la lista la muestra como `22/10/2026`.
3. Editar permite cambiar o borrar la fecha sin alterar el título ni `Completed`.
4. Una fecha mal formada no se guarda: respuesta 422 con mensaje en `due_date`, nunca 500.
5. La migración añade una columna `date` nullable; las tareas existentes quedan sin fecha y la migración es reversible.
6. (Nuevo, UX) Si el alta falla por validación, el campo de fecha conserva el valor enviado en vez de vaciarse.

## Comportamiento actual y propuesta

Estado verificado leyendo el código de la rama `feat/issue-3-add-task-due-date` (commits `d13ec2c` y `7913cf0`, aún no en `main`). **La funcionalidad ya está implementada.** Este plan no la rediseña: verifica que cumple los criterios 1 a 5, añade el criterio 6 y deja explícitas las decisiones pendientes.

Implementado en la rama (comprobado):
- `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`. El `.down.fizz` hace `drop_column("tasks", "due_date")`.
- `models/task.go`: `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`. `Validate` no cambia.
- `actions/tasks.go`:
  - `parseDueDate(raw string) (*time.Time, error)`: cadena vacía → `nil`; si no, `time.Parse("2006-01-02", raw)`.
  - `TasksCreate` y `TasksUpdate` leen `due_date` como `string`, lo parsean y, si falla, añaden `dueDateInvalidMessage` a `due_date` y responden 422.
  - `TasksUpdate` asigna `task.DueDate` en cada PUT (ver R6).
- `templates/tasks/index.plush.html`: `<input type="date" name="due_date">` en el alta (sin `value`, ver R8); muestra `Fecha límite: DD/MM/YYYY` solo si `t.DueDate != nil`.
- `templates/tasks/edit.plush.html`: `<input type="date" name="due_date">` precargado con `YYYY-MM-DD`, con ayuda y error.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`, ambos con variables existentes.
- Tests presentes:
  - `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.
  - `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`.

Lo que **no** está hecho:
- Ninguno de estos tests ni comandos de validación se ha ejecutado en esta rama.
- **Gap R8**: tras un 422 en el alta, `index.plush.html` renderiza el input de fecha vacío. El título sí se conserva. No rompe los criterios 1 a 5, pero es un fallo visible en el flujo de alta con fecha.

Propuesta: (a) ejecutar la validación completa; (b) corregir R8 con el cambio mínimo en `index.plush.html`, con un test de regresión; (c) mantener R6 como está (reemplazo completo en PUT), documentado como decisión.

## Alcance

Incluido:
- Ejecutar la validación de la rama y registrar el resultado.
- Comprobar la reversión de la migración (criterio 5).
- Corregir R8: el input `due_date` del alta muestra el valor enviado tras un 422, en formato `YYYY-MM-DD`.
- Añadir un test de regresión para R8 en `actions/tasks_test.go`.

Excluido:
- Hora, zona horaria, recordatorios o notificaciones.
- Ordenar o filtrar por fecha; resaltar tareas vencidas.
- Cambiar la semántica de PUT sin `due_date` (R6).
- Tocar `TasksToggle`, `TasksDestroy`, rutas, `fixtures/` o el seed.
- Incluir en el PR los archivos ajenos al cambio: `config/adw.yml`, `plans/add-task-due-date.md`, `migrations/schema.sql`, `.claude/skills/ship/` y los planes sin seguimiento `plans/issue-3-cierre-fecha-limite.md`, `plans/issue-3-verificar-fecha-limite.md`, `plans/verificar-fecha-limite-tareas.md`. Este plan (`plans/issue-3-fecha-limite-solo-dia.md`) sí va en el PR.

## Diseño por capas

Capas afectadas:

- **Renderizado** (`templates/tasks/index.plush.html`): añadir `value` al input `due_date` del alta, con el mismo patrón que `edit.plush.html`:
  `value="<%= if (task.DueDate != nil) { %><%= task.DueDate.Format("2006-01-02") %><% } %>"`.
- **Persistencia, modelo, acción, CSS**: ya implementados; sin cambios.
- **JavaScript**: no aplica.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz` | Columna nullable y su reversión |
| `models/task.go` | Campo `DueDate` |
| `actions/tasks.go` | `parseDueDate`, alta y edición con `due_date` (sin cambios) |
| `templates/tasks/index.plush.html` | Alta con fecha (corregir `value`, R8) y lista |
| `templates/tasks/edit.plush.html` | Edición con fecha precargada (sin cambios) |
| `actions/tasks_test.go` | Tests de acción; aquí se añade el de R8 |
| `models/task_test.go` | Tests de modelo (sin cambios) |
| `.github/workflows/ci.yaml` | Solo lectura: orden de validación del CI |

## Implementación paso a paso

1. Confirmar que la rama contiene `d13ec2c` y `7913cf0` y que `git status` no muestra cambios en los archivos de la feature.
2. Preparar la base de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks` sin error.
3. Ejecutar la validación de la sección siguiente. Si algo falla dentro del alcance, corregir solo ese punto.
4. Añadir el test de regresión R8 en `actions/tasks_test.go`, siguiendo el patrón de `Test_TasksCreate_InvalidDueDate`:
   - Nombre sugerido: `Test_TasksCreate_InvalidTitleKeepsDueDate`.
   - Enviar `title=""` y `due_date=2026-10-22` a `POST /tasks`.
   - Esperar 422 y que el HTML contenga `name="due_date"` con `value="2026-10-22"`.
   - Comprobar que el test **falla** antes del cambio de plantilla (si no falla, el test no demuestra R8).
5. Añadir `value` al input `due_date` de `templates/tasks/index.plush.html`, como en el diseño.
6. Comprobar la reversión (criterio 5): `soda migrate down -e test` y después `soda migrate -e test`.
7. No stagear los archivos ajenos listados en Excluido.

## Tests y validación

Comprobaciones pertinentes, en este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración se aplica sin error.
- `go test -p 1 ./...` → todos pasan, incluidos los tests de fecha límite y el nuevo test de R8.
- `go build ./...` → compila.
- Reversión (paso 6): `soda migrate down -e test` seguido de `soda migrate -e test`.

Requisitos de entorno: PostgreSQL 17 accesible con la configuración de `database.yml` para `test` (en CI: `postgres:17.6-alpine`, usuario y contraseña `postgres`, base `todo_test`). En esta máquina el puerto 5432 responde, pero no hay `psql` ni `pg_isready`; la conexión de los tests **no está comprobada**. `soda` está en `~/go/bin/soda`.

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`, `Test_TasksIndex_ListsTasks`.
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`, `Test_TasksUpdate_PendingTaskRemainsPending`.
- Criterio 4 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → paso de migración y reversión.
- Criterio 6 → test nuevo de R8.

Estos tests no se han ejecutado todavía; la validación real es el paso de implementación.

## Riesgos y recuperación

- **R1, binding de fecha**: resuelto; la fecha se lee como cadena y se parsea con `time.Parse`, sin depender de la conversión por defecto de Buffalo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve medianoche UTC. Solo se muestra año/mes/día, así que no hay desplazamiento de día.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas en PostgreSQL. Rollback: `soda migrate down -e <entorno>`.
- **R4, formato en el navegador**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma; `DD/MM/YYYY` es solo presentación.
- **R5, lectura de la issue**: "fechaimie" es ambiguo. Se adopta fecha límite solo de día. Si la intención era otra, revisar antes de abrir el PR.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` asigna `DueDate` en cada petición. El formulario de edición siempre envía el campo, así que la UI no se ve afectada; un cliente que omita el campo sí borrará la fecha. **Decisión**: se mantiene el reemplazo completo (comportamiento actual). Cambiarlo a "conservar si el campo no llega" es una decisión del usuario y queda fuera de este plan.
- **R7, archivos ajenos a la feature**: ya existían sin confirmar o sin seguimiento; no entran en el PR.
- **R8, fecha perdida tras 422 en el alta**: corregido en el paso 5, con test de regresión (paso 4). Es solo UX; si el test no falla antes del cambio, revisar el test.
- **Recuperación**: si la validación falla, revertir la migración con `soda migrate down -e test`, corregir o descartar el cambio concreto. No se modifica ningún dato existente: la columna es nueva y nullable.
