# Feature: Fecha límite opcional de tareas, solo día (issue #3)

## Encargo y criterios de aceptación

Encargo (issue #3, título "añadir fechaimie alas task", cuerpo "vale un dia sin hora"): cada tarea puede tener una fecha límite opcional que indica solo el día, sin hora.

Criterios de aceptación:

1. Crear una tarea con título y, opcionalmente, fecha (`YYYY-MM-DD`). Sin fecha, la tarea queda sin fecha límite.
2. Editar una tarea permite poner, cambiar y quitar la fecha límite.
3. La fecha se guarda y se muestra como día, sin hora ni zona horaria visible.
4. Una fecha no válida se rechaza con un mensaje en el formulario y no se guarda nada.
5. Un error de título conserva la fecha escrita por el usuario en el formulario.
6. Completar o reabrir una tarea conserva su fecha límite.
7. La lista de tareas muestra la fecha límite de cada tarea que la tenga.

Requisitos explícitos (del encargo): fecha límite opcional, solo día sin hora.

Inferencias (no dichas en la issue, tomadas del código existente): formato de entrada `YYYY-MM-DD` del `<input type="date">`; visualización `DD/MM/YYYY` en la lista y `YYYY-MM-DD` en el campo de edición; mensaje de error en español como el resto de la app.

## Comportamiento actual y propuesta

Estado actual en la rama `feat/issue-3-add-task-due-date` (verificado leyendo el código, no ejecutando tests):

- Migración `migrations/20261022120000_add_due_date_to_tasks.up.fizz`: añade `due_date` de tipo `date`, nullable. El down la elimina.
- Modelo `models/task.go`: campo `DueDate *time.Time` con etiquetas `db:"due_date"` y `form:"due_date"`. `Validate` solo valida el título.
- Acciones `actions/tasks.go`: `parseDueDate` convierte `YYYY-MM-DD` en `*time.Time` (vacío = nil). `TasksCreate` y `TasksUpdate` leen `due_date`, añaden el error `due_date` si el parseo falla, y re-renderizan con `dueDateInput` si hay errores. `TasksToggle` no toca la fecha.
- Plantillas: `templates/tasks/index.plush.html` (campo de fecha en el alta, etiqueta "Fecha límite: DD/MM/YYYY" en cada tarea) y `templates/tasks/edit.plush.html` (campo de fecha con ayuda y error).
- CSS `public/assets/app.css`: clases `.task-due` y ajuste de `.new-task-row input[type="date"]`.
- Tests ya presentes: `actions/tasks_test.go` (crear con/sin fecha, fecha inválida en crear y en editar, título inválido conserva fecha, editar pone/quita fecha, prellenado en edición, índice muestra fecha, completar conserva fecha) y `models/task_test.go` (opcional, persiste solo día).

Propuesta: no hay cambio funcional pendiente identificado. El trabajo restante es **validar** que la implementación existente cumple los criterios, ejecutando los tests con la base de datos preparada, y abrir el PR. Si la validación falla, el arreglo debe ser mínimo y dentro de los criterios de arriba.

## Alcance

Incluido:

- Verificación de la implementación existente (migración, modelo, acciones, plantillas, CSS, tests).
- Ejecución de la validación del repositorio.
- Corrección mínima solo si una validación demuestra que un criterio no se cumple.

Excluido:

- Hora de vencimiento, recordatorios, ordenación por fecha, filtros por fecha, zonas horarias.
- Cambios en `config/adw.yml`, `.claude/skills/ship/`, `migrations/schema.sql` (vacío y sin versionar) y los `plans/` no versionados: son cambios previos ajenos a la feature y no entran en el PR.
- Otros cambios de los worktrees bajo `.claude/worktrees/`.

## Diseño por capas

- **Persistencia**: columna `due_date date NULL` en `tasks`, mediante migración fizz (ya creada). Sin índice.
- **Modelo**: `DueDate *time.Time`; la validación del modelo no cambia porque el parseo ocurre en la acción.
- **Acción/ruta**: `parseDueDate` centraliza el parseo; `TasksCreate`, `TasksUpdate` y `TasksEdit` gestionan el valor y el error. Rutas existentes sin cambios.
- **Renderizado**: campos `<input type="date">` en alta y edición, mensajes `errors.Get("due_date")`, visualización en la lista.
- **CSS**: clases `.task-due` y ajuste del campo de fecha en la fila de alta. Sin JavaScript.

## Archivos relevantes

| Ruta | Responsabilidad |
|---|---|
| `migrations/20261022120000_add_due_date_to_tasks.up.fizz` / `.down.fizz` | Añade y elimina `due_date` |
| `models/task.go` | Campo `DueDate` con etiquetas de BD y formulario |
| `models/task_test.go` | Opcionalidad y persistencia solo día |
| `actions/tasks.go` | `parseDueDate`, alta, edición, errores de fecha |
| `actions/tasks_test.go` | Comportamiento HTTP de la fecha límite |
| `actions/actions_test.go` | Helpers de la suite (`authed`, etc.) |
| `templates/tasks/index.plush.html` | Campo de fecha en el alta y visualización en la lista |
| `templates/tasks/edit.plush.html` | Campo de fecha en la edición con ayuda y error |
| `public/assets/app.css` | Estilos de `.task-due` y del campo de fecha |
| `database.yml` | Configuración de la base `test` (`todo_test`) |

## Implementación paso a paso

1. Comprobar que la rama contiene los cambios descritos: `git diff main...HEAD --stat` debe listar los 10 archivos de la sección anterior.
2. Confirmar que los cambios sin confirmar ajenos a la feature (`config/adw.yml`, `.claude/skills/ship/`, `migrations/schema.sql`, `.txt`, `.x`, `plans/` sin versionar) no se añaden al PR.
3. Preparar la base de tests: PostgreSQL accesible en `DB_HOST` (por defecto `127.0.0.1`), base `todo_test` con migraciones aplicadas, incluida `20261022120000`.
4. Ejecutar la validación de la sección siguiente. Si un test falla, leer su salida y decidir si el fallo es del código de la feature o del entorno.
5. Solo si un criterio no se cumple: corregir en el archivo afectado con el patrón existente (`parseDueDate`, `verrs.Add`, `dueDateInput`), añadiendo o ajustando el test que lo demuestra.
6. Confirmar el resultado y preparar el PR con los archivos de la lista anterior.

## Tests y validación

Tests que demuestran el comportamiento (ya existen en la rama):

- `actions/tasks_test.go`: `Test_TasksCreate_WithDueDate`, `Test_TasksCreate_InvalidDueDate`, `Test_TasksCreate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_InvalidTitleKeepsDueDate`, `Test_TasksUpdate_SetsAndClearsDueDate`, `Test_TasksUpdate_InvalidDueDate`, `Test_TasksEdit_PrefillsDueDate`, `Test_TasksIndex_ShowsDueDate`, `Test_TasksToggle_KeepsDueDate`.
- `models/task_test.go`: `Test_Task_DueDate_Optional`, `Test_Task_DueDate_DayOnly_Persists`.

Condiciones necesarias:

- Los tests de `actions/` requieren `GO_ENV=test`. Sin él, la app escribe en `todo_development` y la suite lee `todo_test`, y fallan en cascada.
- Los tests de `models/` y `actions/` requieren PostgreSQL y la migración aplicada en `todo_test`.
- Los tests de `adws/` fallan en Windows independientemente de esta feature; no forman parte de la validación de esta feature.

Comandos de validación:

```
gofmt -l <archivos tocados>
go vet ./...
GO_ENV=test go test -p 1 -count=1 ./models/ ./actions/
go build ./...
```

Nota: `gofmt -l .` en este checkout lista casi todos los archivos Go porque el árbol de trabajo tiene finales de línea CRLF (`core.autocrlf=true`); la copia de `main` no aparece en la lista. Comprobar formato con `gofmt -l` sobre los archivos tocados y tratar la lista completa como ruido de entorno, salvo que `main` también la produzca.

Resultado esperado: `go vet` y `go build` sin salida; `gofmt` sin archivos de la feature; tests de `models/` y `actions/` en verde.

## Riesgos y recuperación

- **Tests no ejecutados todavía**: la rama no tiene un resultado de validación verificado en este plan. Riesgo principal: que un test falle por el entorno, no por el código.
- **Entorno de test**: sin `GO_ENV=test` o sin la migración en `todo_test`, los fallos parecerán bugs de la feature. Recuperación: revisar la variable y la migración antes de tocar código.
- **Zona horaria**: `time.Parse` produce medianoche UTC; el formato `YYYY-MM-DD` se reconstruye sin desplazamiento, por lo que el día no cambia. Si un test lo contradice, corregir en `parseDueDate`.
- **Campos sin versionar**: `migrations/schema.sql` está vacío y sin versionar; no debe añadirse al PR.
- **Alcance**: si la validación revela una necesidad fuera de la lista de criterios, detenerse y consultar; no ampliar la feature.
- **Recuperación general**: los cambios de la feature están en commits de la rama; si algo sale mal, revertir solo el commit afectado, sin tocar los cambios ajenos sin confirmar.
