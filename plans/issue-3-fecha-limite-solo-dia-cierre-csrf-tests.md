# Feature: fecha límite opcional de solo día en las tareas (issue #3)

## Encargo y criterios de aceptación

Fuente: issue #3 de `clerma-arch/desarrollo-agentico-todo-03`, título "añadir fechaimie alas task" (leído como "añadir fecha límite a las tasks"), cuerpo "vale un dia sin hora".

Requisitos explícitos:

1. Cada tarea puede tener una fecha límite opcional.
2. La fecha límite es solo día, sin hora.

Criterios de aceptación (comportamiento observable):

- CA1. Al crear una tarea con `due_date` válida (`YYYY-MM-DD`), la tarea se guarda con esa fecha, sin hora.
- CA2. Al crear una tarea sin `due_date`, la tarea se guarda sin fecha (`NULL`).
- CA3. Un `due_date` que no es `YYYY-MM-DD` válido devuelve 422 con el mensaje "La fecha no es válida.", no guarda la tarea y conserva el texto introducido en el formulario.
- CA4. Editar una tarea muestra su fecha en el campo `date`; cambiarla la actualiza; enviar el campo vacío la borra.
- CA5. Completar o reabrir una tarea (toggle) no altera su fecha límite.
- CA6. La lista muestra la fecha límite de las tareas que la tienen como "Fecha límite: dd/mm/yyyy".
- CA7. Un título inválido en alta o edición mantiene la fecha introducida en el formulario y no persiste cambios.

Inferencias (no pedidas literalmente, derivadas del código existente):

- I1. El formato de visualización en la lista es `dd/mm/yyyy` (ya implementado en `templates/tasks/index.plush.html`).
- I2. No se valida que la fecha sea futura ni que esté dentro de un rango.

## Comportamiento actual y propuesta

Estado de la rama `feat/issue-3-add-task-due-date` (ya tiene commits sobre `main`, entre ellos `d13ec2c feat(tasks): añadir fecha límite opcional (solo día)`):

- Migración `migrations/20261022120000_add_due_date_to_tasks.{up,down}.fizz`: añade `due_date` tipo `date`, nullable.
- `models/task.go`: campo `DueDate *time.Time` con tags `db`/`form`.
- `actions/tasks.go`: `parseDueDate` (formato `2006-01-02`, vacío = nil), usado en `TasksCreate` y `TasksUpdate`; `dueDateInvalidMessage`; `dueDateInput` para repintar el valor enviado en error.
- Templates: el alta de `index` tiene `<input type="date" name="due_date">` y la lista muestra la fecha; `edit` tiene el campo con ayuda "Opcional. Solo indica el día, sin hora." y errores `due-date-error`.
- `public/assets/app.css`: clases `.task-due` y `.new-task-row input[type="date"]`.
- Tests en `actions/tasks_test.go` y `models/task_test.go` que cubren CA1–CA7 (ver sección de tests).

Propuesta: **no hay cambio funcional pendiente.** La funcionalidad cumple CA1–CA7 según el código y los tests. El trabajo restante es verificar, sin cambiar código de aplicación, y abrir el PR con este plan.

### Verificación ejecutada (sesión de redacción)

- `go vet ./...` y `go build ./...`: sin errores.
- `gofmt -l .`: marca casi todos los `.go`, pero es un artefacto de fin de línea: los archivos están en CRLF en el disco (`core.autocrlf=true`, índice en LF) y al quitar los `\r` gofmt no reporta nada. No es un problema de formato del cambio.
- `go test -p 1 ./...` con `GO_ENV=test`:
  - `todo/actions`: **ok**, con todos los tests de fecha límite (CA1–CA7).
  - `todo/models`: **ok**.
  - `todo/adws`: **falla**, pero también falla en `HEAD` (verificado en un worktree temporal). El paquete no se toca en esta rama. Causa: sus tests crean un `gh` falso en bash, y en Windows se invoca el `gh` real contra el repositorio `example/todo`.

### Hallazgo: `GO_ENV=test` es obligatorio

La app elige la base de datos con `GO_ENV` (`models/models.go:18`, `actions/app.go:21`), que por defecto vale `development`. Sin `GO_ENV=test`, las peticiones HTTP de los tests escriben en `todo_development` mientras los `as.DB` leen de `todo_test`. Resultado: 21 fallos en `actions` que no son fallos reales del código. `buffalo test` fija `GO_ENV=test`; `go test` directo no.

### Cambios sin confirmar en la rama (fuera del PR)

Antes de empezar había cambios sin confirmar. Ninguno forma parte de esta funcionalidad:

- `actions/actions_test.go` y `actions/tasks_test.go`: helpers `authed`/`csrfToken` para enviar token CSRF en los tests. **No son necesarios**: `HEAD` pasa los tests de `actions` sin ellos. Se dejan fuera del PR y se conservan sin confirmar.
- `config/adw.yml`: cambio del trigger (repo y prompt). Fuera de alcance.
- `.txt`, `migrations/schema.sql`, `.claude/skills/ship/`, y decenas de archivos `plans/issue-3-*.md` sin seguimiento: fuera de alcance.

## Alcance

Incluido:

- Verificación de CA1–CA7 contra el código y los tests existentes.
- Commit con este plan (`plans/issue-3-fecha-limite-solo-dia-cierre-csrf-tests.md`).
- PR de la rama contra `main`, que incluye los commits de funcionalidad ya existentes.

Excluido:

- Helpers CSRF sin confirmar en `actions/*_test.go` (no necesarios).
- `config/adw.yml`, `.txt`, `migrations/schema.sql` y el resto de planes sin seguimiento.
- Validación de fecha futura, rango de fechas, zona horaria, ordenación o filtros por fecha.
- Cambios en `models/task.go` o `actions/tasks.go` salvo que un test falle por una causa real del código.

## Diseño por capas

- **Persistencia**: migración `20261022120000_add_due_date_to_tasks` (existe). `date` nullable.
- **Modelo** (`models/task.go`): `DueDate *time.Time`. Sin validación de fecha en `Validate`: el parseo está en la acción, como en el resto del repo.
- **Acción** (`actions/tasks.go`): `parseDueDate` + `dueErr` → `verrs.Add("due_date", ...)`, igual en alta y edición.
- **Renderizado** (`templates/tasks/*.plush.html`): campo `date`, `dueDateInput` para repintar, mensajes `due-date-error`.
- **CSS** (`public/assets/app.css`): `.task-due` y `input[type="date"]` dentro de `.new-task-row`.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` | Añade columna `due_date date null` |
| `models/task.go` | Campo `DueDate` |
| `models/task_test.go` | Tests de modelo: opcional y persistencia de día |
| `actions/tasks.go` | `parseDueDate`, alta, edición, toggle, errores |
| `actions/tasks_test.go` | Tests de acción: CA1–CA7 |
| `actions/app.go` | Middleware y selección de entorno (`GO_ENV`) |
| `models/models.go` | Conexión a base de datos según `GO_ENV` |
| `templates/tasks/index.plush.html` | Campo de alta y visualización en lista |
| `templates/tasks/edit.plush.html` | Campo de edición y ayuda |
| `public/assets/app.css` | Estilos de fecha |
| `database.yml` | Conexión de test (PostgreSQL) |

## Implementación paso a paso

1. **Confirmar el estado** con `git status --short --branch` y `git diff main...HEAD --stat`. Comprobar que el código de fecha límite coincide con lo descrito arriba.
2. **No tocar** los helpers CSRF sin confirmar ni `config/adw.yml`. Quedan sin confirmar fuera del PR.
3. **Revisar cobertura** de CA1–CA7 (tests existentes):
   - CA1: `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists`
   - CA2: `Test_Task_DueDate_Optional`, asserts de `Test_TasksCreate_Valid`
   - CA3: `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate`
   - CA4: `Test_TasksEdit_PrefillsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`
   - CA5: `Test_TasksToggle_KeepsDueDate`
   - CA6: `Test_TasksIndex_ShowsDueDate`
   - CA7: `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`

   Si un criterio se queda sin test, añadir **uno** siguiendo el patrón del vecino. No ampliar el alcance.
4. **Validar** (sección siguiente). Si algo falla por causa del código, parar y documentar; no cambiar el alcance.
5. **Commit y PR** solo con el plan (y los commits de la rama ya existentes).

## Tests y validación

Precondiciones:

- PostgreSQL en `127.0.0.1:5432` (el servicio `db` de `compose.yaml`: `docker compose up -d db`), con la base `todo_test` creada.
- **`GO_ENV=test`** en todas las ejecuciones de tests de aplicación.

Comandos, desde la raíz del repo:

```
gofmt -l .                       # ver nota CRLF en "Verificación ejecutada"
go vet ./...
GO_ENV=test go test -count=1 -p 1 ./...
go build ./...
```

Resultado esperado:

- `go vet` y `go build`: sin errores.
- `GO_ENV=test go test`: `todo/actions` y `todo/models` pasan. `todo/adws` falla igual que en `HEAD` (preexistente, fuera de alcance).
- `gofmt -l .`: los únicos archivos marcados son los CRLF ya conocidos; sin CR, limpio.

Criterio de parada: un fallo en `actions` o `models` bloquea el PR. Un fallo en `adws` solo se documenta, porque también ocurre en `HEAD`.

## Riesgos y recuperación

- **Falsos fallos por `GO_ENV`**: sin `GO_ENV=test`, `actions` falla y escribe datos de prueba en `todo_development`.
- **Filas de prueba en `todo_development`**: durante la verificación quedaron 52 filas con títulos de fixtures ("Nueva tarea de prueba", "Entregar el informe") creadas desde las 15:00. No se han borrado; hay que decidir si se eliminan.
- **Preexistente en `adws`**: fallo en Windows con el `gh` falso. No bloquea este PR, pero conviene arreglarlo aparte.
- **Plans duplicados en `plans/`**: la rama acumula decenas de planes casi iguales. No afecta al código; decidir en el PR qué planes se incluyen.
- **Recuperación**: `git diff main...HEAD` muestra los commits de la funcionalidad; son revertibles sin tocar el working tree.
