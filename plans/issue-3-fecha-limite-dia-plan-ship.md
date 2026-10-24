# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Encargo: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora". El título es ambiguo; se adopta la lectura **fecha límite opcional por tarea, con precisión de día y sin hora**.

Requisitos explícitos (cuerpo de la issue):
- La fecha es solo de día, sin hora.

Inferencias (no escritas en la issue):
- El campo es opcional.
- Se indica al crear y al editar.
- Se muestra en la lista y se precarga en el formulario de edición.

Criterios de aceptación:
1. Crear una tarea sin fecha funciona como antes.
2. Crear con `2026-10-31` guarda el día sin hora y la lista lo muestra como `31/10/2026`.
3. Editar cambia o borra la fecha sin alterar el título ni `Completed`.
4. Completar o reabrir una tarea con fecha conserva la fecha.
5. Una fecha mal formada no se guarda: 422 con mensaje en `due_date`, nunca 500.
6. La migración añade una columna nullable y es reversible.

## Comportamiento actual y propuesta

Estado verificado en `feat/issue-3-add-task-due-date`, por delante de `main` con 9 commits (`d13ec2c` … `0125661`). **La funcionalidad ya está implementada en el código committeado.** Este plan no propone diseño nuevo: comprueba los criterios y corrige solo lo que falle.

Implementado (leído en el código):
- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz` (`add_column("tasks", "due_date", "date", {"null": true})`) y su `.down.fizz` (`drop_column`).
- `models/task.go`: `DueDate *time.Time` con `db`, `form` y `json` `due_date`.
- `actions/tasks.go`: `parseDueDate` (vacío → `nil`; si no, `time.Parse("2006-01-02")`). `TasksCreate` y `TasksUpdate` leen `due_date` como cadena y, si falla, añaden `due_date` a los errores y responden 422.
- `templates/tasks/index.plush.html` y `edit.plush.html`: input `type="date"`, lista con `Fecha límite: DD/MM/YYYY` solo si hay fecha.
- `public/assets/app.css`: `.task-due` y `.new-task-row input[type="date"]`.

Los tests de acciones (`as.Nil(task.DueDate)` en create y `Test_TasksToggle_KeepsDueDate`) ya están committeados en `8a4bea5` y `0125661`. No quedan cambios sin confirmar en código de la feature.

Cambios sin confirmar **fuera del PR**:
- `config/adw.yml` (modificado), `migrations/schema.sql`, `.claude/skills/ship/` y planes sin seguimiento distintos de este.

## Alcance

Incluido:
- Verificar los criterios 1 a 6 contra el código committeado.
- Corregir únicamente un fallo detectado en esa verificación.
- Abrir el PR con los commits existentes de la rama más este plan.

Excluido:
- Hora, zona horaria explícita, recurrencia y filtros por fecha.
- Cambiar el comportamiento de PUT sin `due_date` (ver R6).
- Cambiar la política de `TasksUpdate` sin `due_date` (ver R6).
- Merge y push forzado.

## Diseño por capas

- **Persistencia y migraciones**: `due_date date NULL` en `tasks`. Ya existe.
- **Modelo**: `Task.DueDate *time.Time`; `nil` = sin fecha. `Validate` no cambia.
- **Acción y ruta**: sin rutas nuevas. Create y Update parsean la cadena y devuelven 422.
- **Renderizado**: formularios y lista. Ya existen.
- **CSS**: dos reglas que reutilizan `--color-muted`.
- **JavaScript**: no aplica.

## Archivos relevantes

- `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: columna nueva y reversión.
- `models/task.go`, `models/task_test.go`: campo y tests de modelo (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`).
- `actions/tasks.go`: `parseDueDate`, `TasksCreate`, `TasksUpdate`.
- `actions/tasks_test.go`: tests de acciones. La versión committeada cubre create, update, index y edit con fecha. El test de toggle está committeado en `0125661`.
- `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`, `public/assets/app.css`.
- `database.yml`: base `todo_test`, usuario `postgres`, host `127.0.0.1:5432`.

## Implementación paso a paso

1. Confirmar que `git log main..HEAD` contiene la feature y que `git status` solo muestra los cambios ajenos listados.
2. Preparar la base de test: `soda migrate -e test`. Debe aplicar `20261022120000_add_due_date_to_tasks`.
3. Ejecutar la validación. Si falla un criterio, corregir solo ese punto y repetir.
4. Comprobar la reversión (criterio 6): `soda migrate down -e test` y luego `soda migrate -e test`.
5. Stage explícito solo de los archivos de la feature y de este plan; push a `origin` sin forzar.

## Tests y validación

En este orden:
- `gofmt -l .` → sin salida.
- `go vet ./...` → sin errores.
- `soda migrate -e test` → aplica la migración nueva.
- `go test -p 1 ./...` → todo en verde.
- `go build ./...` → compila.
- Reversión de la migración (paso 4).

Cobertura por criterio:
- 1 → `Test_Task_Create_Valid`, `Test_TasksCreate_Valid`.
- 2 → `Test_Task_DueDate_DayOnly_Persists`, `Test_TasksCreate_WithDueDate`, `Test_TasksIndex_ShowsDueDate`.
- 3 → `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_PersistsTitleOnly`.
- 4 → `Test_TasksToggle_KeepsDueDate`.
- 5 → `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
- 6 → paso 4.

Requisitos de entorno: PostgreSQL accesible en `127.0.0.1:5432` con la configuración de `database.yml`. `soda` está en `/c/Users/Cristina/go/bin/soda`. Si PostgreSQL no responde, la validación de base de datos queda **no verificada** y debe informarse así.

## Riesgos y recuperación

- **R1, parseo**: la fecha se lee como cadena y se parsea con `time.Parse`; no depende del binding por defecto de Buffalo.
- **R2, zona horaria**: sin zona, `time.Parse` devuelve medianoche UTC; la presentación solo usa día, mes y año.
- **R3, migración en entornos existentes**: `add_column` nullable no reescribe filas. Reversión: `soda migrate down -e <entorno>`.
- **R4, formato**: `<input type="date">` envía `YYYY-MM-DD`; `DD/MM/YYYY` es solo presentación.
- **R5, lectura de la issue**: "fechaimie" es ambiguo; se adopta fecha límite de solo día. Si la intención era otra, revisar antes de mergear.
- **R6, PUT sin `due_date` borra la fecha**: `TasksUpdate` reasigna `DueDate` en cada petición. El formulario siempre envía el campo, así que la UI no se ve afectada; un cliente que omita el campo borra la fecha. **Decisión pendiente del usuario**: mantener el reemplazo completo (comportamiento actual) o conservar la fecha si el campo no llega. Este plan no lo cambia.
- **R7, tests de la feature**: committeados en `8a4bea5` y `0125661`; no hay tests pendientes fuera del PR.
- **R8, entorno de tests**: `GO_ENV=test` es obligatorio (lo define el CI). Sin él, el middleware CSRF responde 403 a los POST y fallan casi todos los tests de `actions`. El paquete `adws` falla igual en `main` limpio (fallo previo, no causado por esta feature).
- **Recuperación**: si la validación falla, `soda migrate down -e test`, corregir solo el punto concreto. La columna es nueva y nullable, así que no se modifican datos existentes.

Plan relacionado: `plans/issue-3-fecha-limite-dia-plan-final.md`.
