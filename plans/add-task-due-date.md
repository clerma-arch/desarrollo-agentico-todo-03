# Feature: añadir fecha (solo día) a las tareas

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo ("fechaimie"); la lectura adoptada es **añadir a cada tarea una fecha opcional, con precisión de día y sin hora**.

Requisitos explícitos (del cuerpo de la issue):
- La fecha es solo de día (`YYYY-MM-DD`), sin hora.

Inferencias (no escritas en la issue; confirmar en el punto de control):
- El campo es opcional: una tarea puede no tener fecha, como hoy.
- Se puede indicar la fecha al crear y al editar una tarea.
- La fecha se muestra en la lista.

Criterios de aceptación:
1. Crear una tarea sin fecha sigue funcionando igual que hoy.
2. Crear una tarea con fecha `2026-10-22` la guarda como día, sin hora, y la lista la muestra.
3. Editar una tarea permite cambiar o borrar la fecha sin alterar el resto de sus datos (título, completada).
4. Una fecha mal formada no se guarda y la respuesta es de error de validación, no un fallo 500.
5. Las migraciones de la base de datos de test y desarrollo contemplan la nueva columna; las tareas existentes quedan sin fecha.

## Comportamiento actual y propuesta

Actual:
- `models/task.go`: `Task` tiene `ID`, `Title`, `Completed`, `CreatedAt`, `UpdatedAt`. No hay fecha de negocio.
- `migrations/20260901000000_create_tasks.up.fizz`: tabla `tasks` sin columna de fecha.
- `actions/tasks.go`: `TasksCreate` hace `c.Bind(task)` sobre `models.Task`; `TasksUpdate` solo binda `Title` mediante una struct anónima.
- `templates/tasks/index.plush.html`: formulario de alta con solo el campo `title`; cada tarea muestra título, estado y acciones.
- `templates/tasks/edit.plush.html`: formulario con solo el campo `title`.

Propuesta:
- Nueva columna `due_date` de tipo `date`, nullable, en `tasks`.
- `Task.DueDate` como `*time.Time` (nil = sin fecha), con `db:"due_date"` y `form:"due_date"`.
- Formularios de alta y edición con `<input type="date" name="due_date">`.
- La lista muestra la fecha (formato `DD/MM/YYYY`) solo si existe.
- Sin cambios en validación de título ni en rutas.

## Alcance

Incluido:
- Migración que añade `due_date` (up y down).
- Campo en el modelo, con validación solo de formato a través del binding de fecha.
- Alta (`TasksCreate`), edición (`TasksUpdate`) y visualización (`TasksIndex`, plantillas).
- Tests de modelo y de acciones que demuestren los criterios 1 a 5.

Excluido:
- Hora, zona horaria, recordatorios o notificaciones.
- Ordenar o filtrar por fecha; resaltar tareas vencidas.
- Cambios en `TasksToggle`, `TasksDestroy`, rutas, fixtures y seed.
- Cambios de estilo salvo que la validación visual lo exija (ver paso 7).
- Renombrar o reinterpretar el título de la issue más allá de la lectura indicada.

## Diseño por capas

- **Persistencia y migración**: columna nullable; no requiere backfill. Las tareas existentes y el seed no cambian.
- **Modelo** (`models/task.go`): campo nuevo. Sin reglas de validación nuevas salvo, si el binding no rechaza fechas inválidas, una comprobación explícita en `Validate`.
- **Acciones** (`actions/tasks.go`): `TasksCreate` ya binda el modelo, así que el campo entra solo. `TasksUpdate` debe ampliar su struct de entrada con `DueDate` y asignarla; una cadena vacía debe dejar `nil`.
- **Renderizado** (`templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`): inputs de tipo fecha y visualización condicional.
- **CSS** (`public/assets/app.css`): solo si la fecha en la lista no se alinea con los estilos existentes. Preferir reutilizar clases actuales (p. ej. `field-hint` o una clase de texto secundario existente).
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.up.fizz` (nuevo): `add_column("tasks", "due_date", "date", {"null": true})`.
- `migrations/20261022120000_add_due_date_to_tasks.down.fizz` (nuevo): `drop_column("tasks", "due_date")`.
- `models/task.go`: añadir `DueDate *time.Time` a `Task`. `time` ya está importado.
- `models/task_test.go`: tests de modelo nuevos. Sigue el patrón de `Test_Task_Create_Valid` y `Test_Task_Toggle_Persists`.
- `actions/tasks.go`: `TasksUpdate` (struct de entrada) y `TasksCreate` (sin cambios de código si el binding funciona).
- `actions/tasks_test.go`: tests de acciones nuevos. Sigue el patrón de `Test_TasksCreate_Valid` y de los tests de `TasksUpdate` que usan `taskForEditing`.
- `templates/tasks/index.plush.html`: input de fecha en el formulario de alta y muestra de fecha en cada `<li>`.
- `templates/tasks/edit.plush.html`: input de fecha en el formulario de edición.
- `public/assets/app.css`: solo si hace falta (ver paso 7).
- Solo lectura: `migrations/embed.go` (ya embebe `*.fizz`), `.github/workflows/ci.yaml` (pipeline de validación), `fixtures/tasks.toml` (sin cambios, la columna es nullable).

## Implementación paso a paso

1. Línea base: `go test -p 1 ./...` debe pasar antes de tocar nada. Si no pasa, detenerse y reportar.
2. Crear la migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz` con `add_column("tasks", "due_date", "date", {"null": true})` y `migrations/20261022120000_add_due_date_to_tasks.down.fizz` con `drop_column("tasks", "due_date")`. Seguir el formato de `20260901000000_create_tasks.up.fizz`.
3. En `models/task.go`, añadir a `Task`:
   ```go
   DueDate   *time.Time `json:"due_date" db:"due_date" form:"due_date"`
   ```
   Colocarlo entre `Completed` y `CreatedAt`. Mantener el alineado de campos que usa `gofmt`.
4. Verificar cómo Buffalo binda `time.Time` / `*time.Time` desde `2006-01-02` (ver paso 5 y el riesgo R1). Si no convierte el formato de fecha de HTML, añadir en `TasksCreate` y `TasksUpdate` una conversión explícita de la cadena antes de asignar. Un valor vacío debe producir `nil`.
5. En `actions/tasks.go`, `TasksUpdate`: añadir `DueDate` a la struct de entrada (`form:"due_date" json:"due_date"`, tipo `string` para poder detectar vacío) y asignarla al modelo: vacío → `nil`; si no, parsear con `time.Parse("2006-01-02", …)`. Si el parseo falla, añadir error de validación en `due_date` y devolver 422 (mismo patrón que el título). Aplicar la misma lógica en `TasksCreate` si el binding directo no basta.
6. Plantillas:
   - `index.plush.html`: en el formulario de alta, añadir `<input type="date" name="due_date" aria-label="Fecha límite">` junto al título. En cada `<li>` mostrar `t.DueDate` solo si no es nil, con formato `DD/MM/YYYY`.
   - `edit.plush.html`: añadir `<input type="date" id="task-due-date" name="due_date" value="...">` con el valor actual en `YYYY-MM-DD` (vacío si nil). Añadir texto de ayuda con `field-hint`.
7. CSS: solo si la fecha en la lista rompe la maquetación. Si hace falta una clase nueva, añadirla en `public/assets/app.css` reutilizando variables existentes.
8. Tests de modelo (`models/task_test.go`):
   - `Test_Task_DueDate_Optional`: crear sin fecha, `DueDate` nil tras recargar.
   - `Test_Task_DueDate_DayOnly_Persists`: guardar `2026-10-22`, recargar y comprobar año/mes/día y que la hora es 00:00.
9. Tests de acciones (`actions/tasks_test.go`):
   - `Test_TasksCreate_WithDueDate`: POST con `due_date=2026-10-22` → 303 y fecha guardada.
   - `Test_TasksIndex_ShowsDueDate`: tarea con fecha → la respuesta contiene `22/10/2026`.
   - `Test_TasksUpdate_SetsAndClearsDueDate`: PUT con fecha y luego con vacío; título y `Completed` se conservan.
   - `Test_TasksUpdate_InvalidDueDate`: fecha `no-es-fecha` → 422 y mensaje en `due_date`.
10. Aplicar migraciones de test: `soda migrate -e test` (como en CI). Sin esto los tests de base de datos fallan.
11. Validación final (ver siguiente sección).

## Tests y validación

Comprobaciones pertinentes (en este orden):
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → la migración nueva se aplica sin error.
- `go test -p 1 ./...` → todos pasan, incluidos los nuevos.
- `go build ./...` → compila.

Requisitos de entorno: PostgreSQL 17 accesible con la configuración de `database.yml` para `test` (en CI: `postgres:17.6-alpine`, base `todo_test`, usuario y contraseña `postgres`), y `soda` instalado (`go install github.com/gobuffalo/pop/v6/soda@v6.4.1`).

Cobertura por criterio:
- Criterio 1 → `Test_Task_Create_Valid` (existente, debe seguir pasando) y `Test_TasksCreate_Valid`.
- Criterio 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- Criterio 3 → `Test_TasksUpdate_SetsAndClearsDueDate`.
- Criterio 4 → `Test_TasksUpdate_InvalidDueDate`.
- Criterio 5 → `soda migrate -e test` y, si se puede, probar `soda migrate down` y `up` en una base temporal.

Los tests nuevos deben fallar antes del cambio (la columna y el campo no existen). Comprobar que fallan con la línea base si es práctico.

## Riesgos y recuperación

- **R1, binding de fecha**: Buffalo puede no convertir `2026-10-22` a `time.Time` por defecto. Mitigación: paso 4-5, parseo explícito en acciones. Verificar con `Test_TasksCreate_WithDueDate` antes de dar por bueno el binding directo.
- **R2, zona horaria**: `time.Parse` sin zona devuelve UTC medianoche; `pop` puede almacenar `date` sin problema, pero comparar fechas con `time.Local` puede desplazar un día. Mitigación: guardar y comparar solo año/mes/día; mostrar con `Format("02/01/2006")` sobre el valor tal cual se leyó.
- **R3, migración en entornos existentes**: `add_column` con `null: true` no bloquea ni reescribe filas en PostgreSQL. Rollback: `soda migrate down` ejecuta `drop_column`.
- **R4, formato de fecha en el navegador**: `<input type="date">` envía `YYYY-MM-DD` con independencia del idioma; el formato `DD/MM/YYYY` es solo de presentación.
- **R5, alcance de la issue**: la lectura de "fechaimie" es una inferencia. Si la intención era otra (p. ej. fecha de creación visible, o fecha límite con hora), este plan debe revisarse antes de implementar.
- **Recuperación**: si la validación falla tras la migración, revertir con `soda migrate down -e test` y descartar los cambios de código; ningún dato existente se modifica porque la columna es nueva y nullable.
