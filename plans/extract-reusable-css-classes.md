# Chore: extraer estilos repetidos de `app.css` en clases y variables reutilizables

## Encargo y resultado esperado

Encargo: "extraer estilos repetidos en las clases css reutilizables y explicarles sin cambiar el aspecto ni el comportamiento de la aplicación".

Resultado esperado:

- `public/assets/app.css` deja de repetir valores y declaraciones idénticas: los valores repetidos pasan a variables CSS (`:root`) y los bloques de declaraciones repetidos pasan a reglas compartidas (selectores agrupados) con nombre claro.
- Cada variable y cada regla compartida lleva un comentario breve en español que explica **qué** reutiliza y **dónde** se aplica ("explicarlas").
- Aspecto y comportamiento idénticos: mismos estilos calculados en todos los elementos, estados y anchos de pantalla; mismo HTML y mismo JS.

## Estado actual y restricciones

Verificado en el repositorio (rama `main`, árbol limpio):

- Aplicación Go/Buffalo. Único CSS: `public/assets/app.css` (482 líneas, sin comentarios, sin preprocesador, sin bundler, sin `package.json`, sin linter CSS). No hay herramientas de frontend y no se deben añadir.
- Se sirve embebido (`public/embed.go`, `//go:embed *`) y lo enlaza `templates/application.plush.html` con `/assets/app.css?v=3` (cache-busting manual).
- Ya existen variables: `--color-text`, `--color-muted`, `--color-border`, `--color-sidebar`, `--color-primary`, `--color-primary-hover`, `--color-focus`, `--color-danger`, `--font-family`.
- **`public/assets/app.js` depende de nombres de clase** (`.task`, `.task-completed`, `.task-status`, `.task-title`, `.task-actions`, `.toggle-label`, atributos `data-task-toggle` / `data-confirm`). Los templates (`templates/tasks/*.plush.html`, `_flash.plush.html`, `application.plush.html`) usan además `.button`, `.button-primary`, `.icon-action`, `.icon-danger`, `.field-hint`, `.field-error`, `.flash`, `.flash-<clave>`, `.sr-only`, etc. **Ningún selector ni nombre de clase existente puede renombrarse ni eliminarse.**
- Los tests de Go (`actions/*_test.go`) no dependen del CSS; la CI (`.github/workflows/ci.yaml`) ejecuta `gofmt`, `go vet`, migraciones y tests. No hay validación automática de CSS: la equivalencia visual se comprueba a mano con el procedimiento de abajo.
- Repeticiones detectadas (a confirmar al implementar con `grep`):
  - Valores: `border-radius: 6px` (≈10 usos); `44px` como tamaño táctil (`min-height`/`width`/`height` en `.app-brand`, `.nav-link`, `.task-toggle-button`/`.icon-action`, `.add-mark`, `input[type="text"]`, `.button`, `.back-link`, y en media queries `calc(100% - 44px)`); `#ffffff` (fondo de `body`, de `input`, color de `.completion-icon` y de `.button-primary`); `#f7e9e4` (`.nav-link` y `.task-toggle-button:hover .completion-icon`); `120ms ease` (dos transiciones); `0.85rem` y `0.8rem` como tamaños de texto secundario; `font-weight: 600`.
  - Bloques: `display: inline-flex; align-items: center` (`.nav-link`, `.task-toggle-button`/`.icon-action`, `.button`, `.back-link`); `text-decoration: none` junto a `cursor: pointer`/`font: inherit`; el patrón "recuadro de estado" (`padding` + `border-radius: 6px` + `font-size: 0.85rem`) en `.field-error` y `.flash`; `.sr-only, .task-status` ya está agrupado (conservar); `.task-toggle-button, .icon-action` ya comparten base (conservar y documentar).
  - Colores de borde/superficie casi repetidos (`#efefeb`, `#dadad4`, `#f2f2ee`, `#eaeae4`, `#f6f6f2`): son valores distintos y **no se unifican** (cambiaría el aspecto).

Restricciones de seguridad del refactor:

- Un valor solo pasa a variable si es **exactamente el mismo** en todos sus usos y representa el mismo concepto. No "limpiar" colores parecidos.
- Agrupar selectores no puede alterar la cascada: respetar el orden de fuente y la especificidad. Si una regla compartida se coloca antes que las específicas, las específicas siguen ganando; nunca mover una declaración por detrás de otra que la sobrescribe (p. ej. `.nav-link` fija `display:flex`, no `inline-flex`; `.button-primary` sobrescribe `.button`; los `@media` van al final y sobrescriben).
- `@media` y `prefers-reduced-motion` se conservan con los mismos puntos de ruptura (760px, 380px, `hover: hover` + `pointer: fine`) y el mismo orden.
- No se pueden usar variables CSS dentro de condiciones de `@media` (los 760px/380px se quedan literales).
- `calc(100% - 44px)` con variable es válido (`calc(100% - var(--tap-size))`).

## Alcance

Incluido:

- Refactor de `public/assets/app.css`: nuevas variables en `:root`, reglas compartidas por selector agrupado, comentarios explicativos, reordenación mínima de reglas agrupadas.
- Subir el parámetro de caché en `templates/application.plush.html` de `app.css?v=3` a `app.css?v=4` para que los navegadores no sirvan la hoja antigua. Es el único cambio en templates.

Excluido:

- Renombrar o eliminar clases/selectores; cambiar HTML, `app.js`, Go, tests, migraciones, Dockerfile o dependencias.
- Cambiar cualquier valor visual (colores, tamaños, espaciados, tipografía, transiciones), unificar valores "parecidos" o rediseñar.
- Añadir herramientas (Sass, PostCSS, stylelint, Node) o crear clases nuevas que requieran editar los templates.
- Reorganizar la hoja en varios archivos.

## Archivos relevantes

| Ruta | Responsabilidad |
| --- | --- |
| `public/assets/app.css` | Única hoja de estilos; archivo a refactorizar. |
| `templates/application.plush.html` | Enlaza la hoja con `?v=3`; subir a `?v=4`. |
| `templates/tasks/index.plush.html`, `templates/tasks/edit.plush.html`, `templates/_flash.plush.html` | Consumidores de las clases; solo lectura, referencia para no romper nombres. |
| `public/assets/app.js` | Depende de nombres de clase (`.task`, `.task-completed`, `.task-status`, `.task-title`, `.task-actions`, `.toggle-label`); solo lectura. |
| `public/embed.go` | Embebe `public/`; sin cambios. |
| `actions/tasks_test.go`, `actions/actions_test.go` | Tests existentes; se ejecutan como regresión, sin cambios. |

## Implementación paso a paso

1. **Línea base (antes de tocar nada).** Comprobar `git status --short --branch` limpio y arrancar la app (skill `env:start` o el comando documentado en el repo; necesita la base de datos de `compose.yaml`). Crear 3 tareas (una larga, una completada) para ver todos los estados, y una vista de edición con error (título vacío) y la lista vacía.
2. **Capturar los estilos calculados de referencia.** En el navegador (Browser pane o `playwright-cli`), para cada página (`/`, `/tasks/<id>/edit`, edición con error, lista vacía, mensaje flash tras crear) y cada ancho (1280, 760, 759, 380, 379 px), ejecutar un script que recorra `document.querySelectorAll('*')` y vuelque `getComputedStyle` de todas las propiedades por elemento (clave: ruta de selector/índice) a JSON. Guardarlo fuera del repo, en el scratchpad (p. ej. `baseline-<pagina>-<ancho>.json`). Añadir captura de pantalla por página/ancho. Para estados `:hover`/`:focus-visible`, usar `playwright-cli` o el navegador: pasar el ratón sobre `.nav-link`, `.task-toggle-button`, `.icon-action`, `.icon-danger`, `.button`, `.button-primary`, `.back-link` y tabular a un enlace/botón/input, y volcar los estilos del elemento. Verificar también el modo `prefers-reduced-motion` (emulado) y `hover: none` (emulación táctil con el preset móvil).
3. **Inventario de repeticiones.** Con Grep sobre `public/assets/app.css` contar ocurrencias de `6px`, `44px`, `#ffffff`, `#f7e9e4`, `120ms ease`, `0.85rem`, `0.8rem`, `font-weight: 600`, `inline-flex`, `text-decoration: none`. Quedarse solo con los valores idénticos usados con el mismo significado.
4. **Variables en `:root`** (añadirlas al bloque existente, con un comentario por grupo, sin tocar las actuales):
   - `--radius: 6px` (radio de esquina de controles, avisos y recuadros).
   - `--tap-size: 44px` (tamaño mínimo de objetivo táctil/ alto de controles).
   - `--color-surface: #ffffff` (fondo de página/campos y texto sobre fondos de color).
   - `--color-primary-soft: #f7e9e4` (fondo suave del color primario: enlace activo y hover del icono de completar).
   - `--transition-fast: 120ms ease` (solo si se usa tal cual en `transition: background-color var(--transition-fast), ...`; si la sintaxis complica la lectura, dejar literal).
   - Sustituir cada uso por `var(...)`. No tocar `#f2ded6`, `#efefeb`, etc., que no se repiten.
5. **Reglas compartidas.** Sin cambiar el nombre de ninguna clase:
   - Agrupar el patrón "control en línea centrado" (`display: inline-flex; align-items: center; min-height: var(--tap-size); text-decoration: none`) para `.back-link`, `.button` y, donde no choque con `display: flex`, los demás. Para `.nav-link` (usa `display: flex`) compartir solo las declaraciones idénticas (`align-items`, `min-height`, `text-decoration`).
   - Agrupar el patrón "caja de aviso" (`border-radius: var(--radius); font-size: 0.85rem`) para `.field-error` y `.flash`, manteniendo `padding`, colores y márgenes propios en cada regla.
   - Mantener `.task-toggle-button, .icon-action` y `.sr-only, .task-status` como están, solo añadiendo comentario.
   - Colocar cada regla compartida **en la posición de la primera regla que la usa** y comprobar que ninguna regla posterior anula por orden algo que antes ganaba la propia (revisar `.button` → `.button:hover` → `.button-primary`).
6. **Comentarios explicativos.** Poner un comentario `/* … */` en español, de una línea (dos como máximo), sobre cada variable nueva, cada regla compartida y cada sección grande (layout, sidebar, lista de tareas, formularios, botones, avisos, utilidades, responsive). Explicar el propósito, no repetir la declaración. Sin comentarios triviales.
7. **Cache-busting.** Cambiar `app.css?v=3` por `app.css?v=4` en `templates/application.plush.html`.
8. **Revisión del diff.** `git diff --stat` debe mostrar solo `public/assets/app.css` y `templates/application.plush.html`. Revisar a mano que no queda ninguna declaración perdida ni desplazada de regla (contar declaraciones por selector antes/después con un volcado simple si hace falta).

## Tests y validación

Comandos (los de la CI; no hay validación de CSS automatizada):

```bash
gofmt -l .
go vet ./...
go build ./...
go test -p 1 ./...
```

(`go test` necesita Postgres según `compose.yaml` y `GO_ENV=test`, como en CI.) Deben dar el mismo resultado que antes de empezar; el CSS no afecta a los tests, pero confirma que el embebido y los templates siguen compilando.

Equivalencia visual y de comportamiento:

1. Con la app arrancada tras el cambio, repetir **exactamente** el volcado del paso 2 de la implementación (mismas páginas, anchos y estados) y compararlo con la línea base (`diff` o comparación JSON). **El resultado esperado es cero diferencias** en estilos calculados; cualquier diferencia es un fallo del refactor, no un cambio aceptado. Forzar recarga sin caché (el `?v=4` ya lo garantiza) y comprobar en red que se sirve la hoja nueva.
2. Comparar visualmente las capturas antes/después (mismos anchos 1280/760/380).
3. Comprobar manualmente el comportamiento: añadir tarea, completar/reabrir (animación del icono y tachado con JS), editar con y sin error, eliminar con confirmación `data-confirm`, foco con teclado (anillo `--color-focus`), hover de acciones en escritorio y acciones siempre visibles en ≤760px.
4. Sin errores 404 ni de consola en el navegador.

Si todo coincide: la tarea está terminada. Reportar el número de líneas/declaraciones repetidas eliminadas.

## Riesgos y recuperación

- **Cambio sutil de cascada** al mover reglas agrupadas (p. ej. que `.button` pase a ganar o perder frente a `.button-primary`/`.button:hover`). Mitigación: agrupar sin reordenar por encima de reglas que sobrescriban y verificar con el volcado de estilos calculados, que lo detecta.
- **Valores "casi iguales" unificados por error** (p. ej. `#f2f2ee` vs `#f6f6f2`). Mitigación: solo variables para valores idénticos; el diff de estilos calculados lo pondría en evidencia.
- **Variables en lugar no permitido** (media queries, `calc` mal formado): dejar el valor literal.
- **Caché del navegador** con CSS antiguo y HTML nuevo: cubierto por `?v=4`.
- **Romper el JS por renombrar clases:** prohibido en el alcance; revisar con Grep que los selectores de `app.js` siguen existiendo en `app.css`/templates.
- Recuperación: el cambio afecta a dos archivos; `git restore public/assets/app.css templates/application.plush.html` (o `git revert` del commit) devuelve el estado anterior sin migraciones ni efectos operativos.
- Incertidumbre: no hay herramienta automática de regresión visual en el repo; la garantía depende del volcado de estilos calculados y de las capturas descritas arriba.
