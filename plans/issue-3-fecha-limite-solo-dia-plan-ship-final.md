# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task" (entendido como "añadir fecha límite a las tareas"), cuerpo: "vale un dia sin hora".

**Requisitos explícitos (del cuerpo de la issue):**

- CA1. Cada tarea puede tener una fecha límite.
- CA2. La fecha límite es solo día: no guarda hora.

**Inferencias (no escritas en la issue, deducidas del código y del patrón del repositorio):**

- CA3. La fecha es opcional: una tarea sin fecha sigue siendo válida (inferido de "vale", que se entiende como "es válido" un día sin hora y, por extensión, sin fecha).
- CA4. Se puede fijar, cambiar y borrar la fecha al crear y al editar la tarea.
- CA5. Una fecha inválida (p. ej. `no-es-fecha` o `2026-02-30`) muestra error de formulario y no guarda la tarea.
- CA6. Al completar o reabrir una tarea la fecha se conserva.
- CA7. La lista muestra la fecha límite solo si existe, en formato `dd/mm/aaaa` (formato español, coherente con la UI en castellano).

## Comportamiento actual y propuesta

**Estado actual en la rama `feat/issue-3-add-task-due-date`** (ya commiteado, respecto a `main`, 10 archivos de código, +272/−31):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, y su `down` con `drop_column`.
- `models/task.go`: campo `DueDate *time.Time` con etiquetas `db`, `form` y `json`.
- `actions/tasks.go`:
  - `parseDueDate(raw string)` convierte `YYYY-MM-DD` en `*time.Time` (UTC, sin hora) y devuelve `nil` para cadena vacía.
  - `TasksCreate` y `TasksUpdate` leen `due_date`, asignan la fecha y añaden el error `due_date` si el parseo falla. Al re-renderizar con 422 conservan el texto introducido (`dueDateInput`).
  - `TasksEdit`, `TasksIndex` inicializan `dueDateInput` a `""`.
- `templates/tasks/index.plush.html`: input `type="date"` en el formulario de creación; muestra `Fecha límite: dd/mm/aaaa` por tarea si `DueDate != nil`; muestra el error `due_date`.
- `templates/tasks/edit.plush.html`: campo de fecha con ayuda "Opcional. Solo indica el día, sin hora.", y error `due_date`.
- `public/assets/app.css`: clases `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests: `actions/tasks_test.go` (creación, edición, fecha inválida, conservación al completar, render de índice y edición) y `models/task_test.go` (`Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`).

**Conclusión de la lectura:** la funcionalidad está implementada en la rama. No se ha identificado ningún requisito de CA1–CA7 sin cubrir en el código. El trabajo pendiente es **validar y cerrar**, no implementar.

**Propuesta:** no añadir código de aplicación. Ejecutar la validación de los tests existentes con la base de test preparada y, solo si falla algo, corregirlo dentro del alcance de la issue.

**Estado del árbol de trabajo (no forma parte de la funcionalidad):**

- `config/adw.yml` modificado (cambia `repo` y `prompt` del trigger). No pertenece a la issue.
- Sin rastrear: `.claude/skills/ship/`, `.txt`, `.x`, `migrations/schema.sql` y más de cien `plans/issue-3-*.md` y `plans/verificar-fecha-limite-tareas.md`. Ninguno debe entrar en el PR de la funcionalidad.

## Alcance

**Incluido:**

- Verificar CA1–CA7 contra el código y los tests existentes.
- Ejecutar las validaciones de la skill `ship` y registrar el resultado.
- Añadir únicamente tests que falten para un CA concreto, si la verificación encuentra un hueco.

**Excluido:**

- Validar que la fecha no sea anterior a hoy (no pedido).
- Recordatorios, ordenación o filtrado por fecha límite.
- Hora, zona horaria o formatos distintos de `YYYY-MM-DD` en la API.
- Cambiar `config/adw.yml`, borrar los planes existentes o tocar los archivos sin rastrear ajenos a la funcionalidad.
- Fusionar el PR o hacer push forzado.

## Diseño por capas

Solo las capas afectadas por lo ya implementado:

- **Persistencia:** columna `due_date date NULL` en `tasks` mediante migración fizz (ya existe; no se cambia).
- **Modelo:** `Task.DueDate *time.Time`. Un `nil` representa "sin fecha".
- **Acción y ruta:** `TasksCreate` (`POST /tasks`) y `TasksUpdate` (`PUT /tasks/{task_id}`) parsean `due_date` con `parseDueDate`. Las rutas no cambian.
- **Renderizado:** plantillas Plush de índice y edición. Las fechas se muestran con `Format`.
- **CSS:** solo estilos de presentación de la fecha; no hay JavaScript.

No hay capa de API JSON nueva: el campo `json:"due_date"` ya existe en el modelo, pero las rutas que devuelven JSON no se han tocado.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y quita la columna `due_date`. |
| `models/task.go` | Campo `DueDate`. |
| `models/task_test.go` | Tests de modelo: opcional y solo día. |
| `actions/tasks.go` | `parseDueDate`, creación y edición con validación de fecha. |
| `actions/tasks_test.go` | Tests de acción y de render. |
| `actions/app.go` | Rutas `POST /tasks` y `PUT /tasks/{task_id}` (sin cambios). |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y fecha visible en la lista. |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición. |
| `public/assets/app.css` | Estilos `.task-due` y del input de fecha. |
| `database.yml` | Conexión `test` a `todo_test` en `127.0.0.1`. |
| `models/models.go` | Selecciona la conexión según `GO_ENV`. |

## Implementación paso a paso

1. Confirmar que la rama está limpia en lo relativo a la funcionalidad: `git log main..HEAD --oneline` debe listar los commits de issue #3 y `git diff main...HEAD -- . ':!plans'` debe coincidir con los 10 archivos anteriores.
2. Verificar CA1–CA7 leyendo los tests:
   - CA1–CA2: `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`.
   - CA3: `Test_TasksCreate_Valid` (`as.Nil(task.DueDate)`), `Test_Task_DueDate_Optional`.
   - CA4: `Test_TasksUpdate_SetsAndClearsDueDate`.
   - CA5: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`.
   - CA6: `Test_TasksToggle_KeepsDueDate`.
   - CA7: `Test_TasksIndex_ShowsDueDate`, `Test_TasksEdit_PrefillsDueDate`.
3. Si algún CA no tiene test, añadir **un** test siguiendo el patrón de `actions/tasks_test.go` (`as.authed(...)`, `as.DB.First`, `as.Contains`). No modificar código de aplicación salvo que un test nuevo falle.
4. Ejecutar las validaciones (ver siguiente sección) con `GO_ENV=test`.
5. Si todo pasa, la rama está lista para el PR existente (#5). No crear un PR nuevo.

## Tests y validación

**Condiciones previas:**

- PostgreSQL escuchando en `127.0.0.1:5432` (comprobado: el puerto responde).
- Base `todo_test` con las migraciones aplicadas.
- **`GO_ENV=test` obligatorio** para `go test ./actions/` y `./models/`. Sin él, `models.DB` apunta a `todo_development` y la suite lee `todo_test`, y los tests fallan en cascada (ver memoria `go-tests-need-go-env-test`).

**Comandos:**

```sh
gofmt -l .
go vet ./...
GO_ENV=test go test -p 1 -count=1 ./models/ ./actions/
go build ./...
```

Criterio de éxito: `gofmt -l .` no lista archivos del cambio; `go vet` y `go build` sin errores; los tests de `models/` y `actions/` en verde.

**Nota:** `go test -p 1 ./...` completo incluye `adws/`, cuyos tests fallan en Windows por el `gh` real (también en `main`). Se registra como fallo preexistente y no bloquea el cambio si `models/` y `actions/` pasan.

## Riesgos y recuperación

- **Falso verde por base incorrecta:** ejecutar sin `GO_ENV=test` puede escribir en `todo_development` y leer datos ajenos. Mitigación: exportar `GO_ENV=test` en el mismo comando.
- **Migración ya aplicada en `todo_development`:** la migración es idempotente por versión de pop; si hace falta revertir, usar el `down.fizz` correspondiente.
- **Formato de visualización:** `dd/mm/aaaa` en la lista y `aaaa-mm-dd` en el input de edición son decisiones de presentación inferidas. Si la issue requiere otro formato, cambiar solo la plantilla.
- **Archivos sin rastrear en el PR:** `config/adw.yml`, `migrations/schema.sql`, `.txt`, `.x`, `.claude/skills/ship/` y los planes no deben entrar en el commit de la funcionalidad.
- **Recuperación:** si la verificación revela un fallo, el cambio se limita a una corrección en `actions/tasks.go` o en la plantilla correspondiente, con su test. Rollback: `git revert` del commit concreto; no reescribir historia.

---

Siguiente paso: `/implement plans/issue-3-fecha-limite-solo-dia-plan-ship-final.md` (en esta situación, la fase de implementación se limita a la verificación y al cierre).
