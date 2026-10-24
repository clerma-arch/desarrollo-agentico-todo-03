# Feature: Fecha límite opcional (solo día) en las tareas — verificación y cierre (issue #3)

## Encargo y criterios de aceptación

Encargo explícito de la issue #3 (`añadir fechaimie alas task`, cuerpo: `vale un dia sin hora`):

- Las tareas pueden tener una fecha límite.
- La fecha es opcional.
- La fecha es solo día, sin hora.

Criterios de aceptación derivados (inferencias marcadas como tales):

1. Crear una tarea sin fecha la deja sin fecha (`due_date` NULL). Explícito en la idea de "opcional".
2. Crear una tarea con fecha `YYYY-MM-DD` la guarda como día. Explícito.
3. Una fecha con formato inválido no crea la tarea y muestra un error en el formulario. Inferencia.
4. La lista muestra la fecha como `dd/mm/aaaa`, sin hora. Inferencia de la UI del proyecto.
5. Editar permite poner, cambiar y quitar la fecha. Inferencia (coherente con "opcional").
6. Completar o reabrir una tarea conserva su fecha. Inferencia.
7. Un error en el título conserva la fecha introducida al volver a mostrar el formulario. Inferencia de los patrones existentes.

## Comportamiento actual y propuesta

Estado del código en la rama `feat/issue-3-add-task-due-date` (verificado en esta sesión):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: `add_column("tasks", "due_date", "date", {"null": true})`, con su `down`.
- Modelo `models/task.go`: campo `DueDate *time.Time` con `db:"due_date"` y `form:"due_date"`.
- Acciones `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en día sin hora y devuelve `nil` si la cadena está vacía; `TasksCreate` y `TasksUpdate` leen `due_date`, validan y reenvían el valor al formulario en caso de error; `TasksEdit` y `TasksIndex` fijan `dueDateInput`.
- Plantillas: `templates/tasks/index.plush.html` (campo de fecha en el formulario de alta y muestra de la fecha en cada tarea) y `templates/tasks/edit.plush.html` (campo con etiqueta, ayuda y error).
- Estilos: `public/assets/app.css` (`.task-due` y ancho del input de fecha).
- Tests: `actions/tasks_test.go`, `actions/actions_test.go` (helper `authed` con token CSRF), `models/task_test.go`.

Propuesta: **no hay cambio de funcionalidad pendiente**. El trabajo consiste en confirmar que la rama cumple los criterios, dejar limpio el alcance del PR y actualizar el PR #5 existente.

## Alcance

Incluido:

- Verificar los criterios anteriores contra el código y los tests de la rama.
- Actualizar el PR #5 (`https://github.com/clerma-arch/desarrollo-agentico-todo-03/pull/5`) con esta verificación si hace falta.

Excluido:

- `config/adw.yml` (cambio local ajeno a la issue).
- `.txt` y `.x` (salidas sueltas de tests fallidos).
- `migrations/schema.sql` (artefacto no versionado).
- `.claude/skills/ship/` (skill, no código de la aplicación).
- Los archivos `plans/*.md` antiguos sin confirmar: se dejan fuera del commit salvo este plan.
- Cambiar los finales de línea de los archivos del repositorio.
- Cualquier cambio de API JSON: el `json:"due_date"` del modelo serializa como timestamp; las rutas son solo HTML y no se toca.

## Diseño por capas

Solo se verifican las capas afectadas:

- Persistencia: migración `date` nullable. Sin cambios.
- Modelo: `DueDate *time.Time`. Sin cambios.
- Acción y ruta: `TasksCreate`, `TasksUpdate`, `TasksEdit`, `TasksIndex`. Sin cambios salvo verificación.
- Render: plantillas Plush. Sin cambios salvo verificación.
- CSS: sin cambios.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y retira la columna `due_date` |
| `models/task.go` | Campo `DueDate` |
| `models/task_test.go` | Fecha opcional y persistencia solo de día |
| `actions/tasks.go` | `parseDueDate`, alta, edición, índice |
| `actions/tasks_test.go` | Tests de alta, edición, error, toggle y render |
| `actions/actions_test.go` | Helper `authed` y `csrfToken` |
| `templates/tasks/index.plush.html` | Campo de fecha en alta y fecha mostrada |
| `templates/tasks/edit.plush.html` | Campo de fecha en edición |
| `public/assets/app.css` | `.task-due` y ancho del input de fecha |

## Implementación paso a paso

1. Confirmar que `git status` solo muestra los archivos excluidos arriba como cambios ajenos.
2. Ejecutar los tests de `actions` y `models` con `GO_ENV=test` (ya hecho en esta sesión: pasan).
3. Revisar que cada criterio de aceptación tiene un test o una comprobación de plantilla (ver tabla de validación).
4. Si algún criterio no tuviera cobertura, añadir solo ese test siguiendo los patrones de `tasks_test.go`. Según el análisis actual, no hace falta.
5. Decidir sobre el formato de finales de línea (ver riesgos). Recomendación: no tocarlo en este PR.
6. Hacer commit solo con los archivos del cambio y, si se quiere, este plan. Mensaje Conventional Commits, por ejemplo `chore(tasks): verificar la fecha límite solo día (issue #3)`.
7. Hacer push a `origin feat/issue-3-add-task-due-date` (sin `--force`).
8. Actualizar la descripción del PR #5 con el resumen, la ruta de este plan y el resultado de la validación.

## Tests y validación

Tests ya presentes en la rama que cubren los criterios:

| Criterio | Test |
|---|---|
| 1 Sin fecha | `Test_TasksCreate_Valid` (`as.Nil(task.DueDate)`), `Test_Task_DueDate_Optional` |
| 2 Fecha guardada como día | `Test_TasksCreate_WithDueDate`, `Test_Task_DueDate_DayOnly_Persists` |
| 3 Formato inválido | `Test_TasksCreate_InvalidDueDate`, `Test_TasksUpdate_InvalidDueDate` |
| 4 Muestra dd/mm/aaaa | `Test_TasksIndex_ShowsDueDate` |
| 5 Poner, cambiar y quitar | `Test_TasksUpdate_SetsAndClearsDueDate` |
| 6 Completar conserva fecha | `Test_TasksToggle_KeepsDueDate` |
| 7 Conserva fecha al fallar título | `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate` |
| Precarga en edición | `Test_TasksEdit_PrefillsDueDate` |

Comandos y resultado observado en esta sesión:

- `GO_ENV=test go test -count=1 -p 1 -v ./actions/ ./models/`: **PASS** (Test_ActionSuite, Test_ModelSuite).
- `go vet ./...`: **sin salida, código 0**.
- `go build ./...`: **sin salida, código 0**.
- `gofmt -l .`: **marca todos los archivos de la rama** por CRLF en el working tree (`git ls-files --eol`: índice `i/lf`, copia de trabajo `w/crlf`). No es un problema de código; ver riesgos.

Comando de validación para el PR:

```
GO_ENV=test go test -count=1 -p 1 ./...
go vet ./...
go build ./...
```

## Riesgos y recuperación

- **gofmt en Windows.** Con `core.autocrlf=true`, la copia de trabajo tiene CRLF y `gofmt -l` lo reporta, aunque el blob del índice es LF y está limpio. Reformatear rompería el diff con el resto del repositorio. Recomendación: no tocar finales de línea; documentar en el PR.
- **Tests de `.txt`.** El archivo `.txt` contiene fallos antiguos de `Test_TasksCreate_*` y de `OnlySpaces`. Son de una ejecución previa; la ejecución actual pasa. No se incluye en el PR. Recuperación: borrar el archivo a mano si el usuario lo confirma.
- **Migración en producción.** `add_column` nullable no reescribe la tabla ni rompe filas existentes. Recuperación: `down.fizz` la elimina.
- **Zona horaria.** `time.Parse` produce medianoche UTC; el formato de salida es solo fecha, por lo que no hay desfase visible. Si en el futuro se muestra con `time.Local`, revisar.
- **Historial de planes.** La rama tiene decenas de planes `plans/issue-3-*` sin confirmar. No se añaden al PR salvo este.

Siguiente paso: `/implement plans/issue-3-fecha-limite-dia-verificado-y-cierre.md` (la implementación se limita a los pasos de cierre; no hay código que cambiar salvo que la verificación encuentre un hueco).
