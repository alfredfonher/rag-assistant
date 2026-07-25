# PRD — UI web profesional para RAG diseñada con Google Stitch

**Producto:** Interfaz web para un RAG local con backend en Go  
**Documento:** Product Requirements Document orientado a Google Stitch  
**Herramienta principal de diseño:** Google Stitch  
**Backend:** API RAG en Go  
**Frontend previsto:** aplicación web responsive generada y refinada desde Stitch  
**Estado:** Propuesta de diseño y construcción  

---

## 1. Propósito del documento

Este PRD define cómo diseñar en Google Stitch una interfaz web profesional para un sistema RAG local cuyo backend está construido en Go.

El documento está preparado para utilizarse como guía directa de trabajo dentro de Stitch. No se limita a describir funcionalidades: organiza el producto en pantallas, componentes, estados, flujos y prompts independientes que pueden introducirse progresivamente en Stitch para generar, comparar y refinar la interfaz.

La aplicación resultante debe permitir:

- conversar con agentes especializados;
- seleccionar colecciones documentales;
- recibir respuestas por streaming;
- consultar citas y fuentes;
- inspeccionar los fragmentos recuperados;
- gestionar documentos e indexación;
- configurar agentes;
- revisar conversaciones anteriores;
- comprobar el estado del backend y de los modelos.

---

## 2. Contexto de uso de Google Stitch

Google Stitch se utilizará como herramienta de diseño generativo para:

1. Crear la dirección visual inicial mediante prompts.
2. Generar las pantallas principales por separado.
3. Explorar variantes de layout y densidad.
4. Refinar pantallas mediante conversación.
5. Utilizar capturas o wireframes como referencias visuales cuando sea necesario.
6. Exportar el resultado a código frontend o trasladarlo a Figma para una fase adicional de diseño.

El proyecto no debe generarse mediante un único prompt gigantesco. El flujo recomendado es incremental:

```text
Definir sistema visual
        ↓
Generar shell principal
        ↓
Generar pantalla de chat
        ↓
Generar fuentes e inspector
        ↓
Generar documentos y colecciones
        ↓
Generar agentes y configuración
        ↓
Generar historial y diagnóstico
        ↓
Refinar estados responsive
        ↓
Exportar y conectar al backend Go
```

---

## 3. Visión del producto

Crear una interfaz de conocimiento profesional, rápida y verificable para consultar información privada mediante un RAG local.

La aplicación debe sentirse como una combinación de:

- cliente de chat moderno;
- espacio de investigación documental;
- consola de administración de conocimiento;
- panel de diagnóstico del sistema.

No debe sentirse como un chatbot genérico ni como una consola técnica improvisada.

### Promesa principal

> Formular una pregunta, obtener una respuesta estructurada y verificar inmediatamente qué documentos respaldan cada afirmación.

---

## 4. Objetivos del producto

### 4.1 Objetivo principal

Permitir que un usuario consulte, verifique y administre un RAG local desde una interfaz web clara y profesional.

### 4.2 Objetivos específicos

- Seleccionar un agente antes o durante una conversación.
- Seleccionar una o varias colecciones documentales.
- Mostrar respuestas progresivamente mediante streaming.
- Visualizar citas vinculadas a fragmentos concretos.
- Abrir el documento, página o sección de origen.
- Gestionar documentos e indexaciones.
- Crear y editar agentes sin cambiar código.
- Mantener historial de conversaciones.
- Mostrar estados operativos comprensibles.
- Separar controles básicos y avanzados.
- Funcionar correctamente en escritorio, tableta y móvil.
- Facilitar el trabajo posterior del desarrollador al exportar el diseño.

---

## 5. Usuarios objetivo

### 5.1 Usuario técnico

Desarrollador o investigador que ejecuta el RAG localmente.

Necesita:

- control de agentes y modelos;
- métricas de recuperación;
- trazabilidad;
- parámetros avanzados;
- diagnóstico del backend.

### 5.2 Usuario no técnico

Persona que consulta una base documental especializada.

Necesita:

- chat fácil de usar;
- fuentes claras;
- mensajes sencillos;
- errores accionables;
- poca exposición a conceptos técnicos.

### 5.3 Administrador documental

Persona responsable de cargar, clasificar y reindexar documentos.

Necesita:

- carga masiva;
- filtros;
- estados de procesamiento;
- manejo de errores;
- metadatos y colecciones.

---

## 6. Principios de diseño

### 6.1 Profesional, no ornamental

El diseño debe ser sobrio, limpio y orientado al trabajo prolongado.

### 6.2 La respuesta ocupa el centro

El chat y el contenido recuperado deben dominar la jerarquía visual.

### 6.3 Las fuentes siempre están disponibles

La trazabilidad no debe quedar oculta detrás de múltiples clics.

### 6.4 La complejidad se revela progresivamente

Los controles avanzados deben estar disponibles sin sobrecargar la experiencia inicial.

### 6.5 Estados explícitos

Toda operación importante debe mostrar su estado:

- conectando;
- recuperando;
- generando;
- validando;
- completado;
- cancelado;
- error;
- reintentando.

### 6.6 Consistencia entre pantallas

Stitch debe mantener el mismo sistema visual, navegación, espaciado, tipografía, colores, radios y tratamiento de componentes en todas las generaciones.

---

## 7. Dirección visual para Stitch

### 7.1 Personalidad visual

- Producto de IA profesional.
- Sobrio y técnico, pero accesible.
- Alta densidad controlada en escritorio.
- Espacios bien definidos.
- Contraste suficiente para sesiones largas.
- Sin efectos excesivos ni estética futurista exagerada.

### 7.2 Estilo recomendado

- Diseño de aplicación SaaS moderna.
- Sidebar izquierda persistente en escritorio.
- Área de trabajo central flexible.
- Panel derecho contextual para fuentes o detalles.
- Bordes discretos.
- Elevación mínima.
- Iconografía lineal y consistente.
- Estados activos claramente diferenciados.

### 7.3 Tema

Diseñar inicialmente en modo oscuro profesional, pero definir también equivalencia para modo claro.

**Modo oscuro sugerido:**

- fondo principal gris carbón profundo;
- superficies ligeramente más claras;
- texto principal blanco cálido;
- texto secundario gris neutro;
- acento azul verdoso o índigo sobrio;
- éxito verde discreto;
- advertencia ámbar;
- error rojo moderado.

### 7.4 Tipografía

- Sans serif moderna y legible.
- Excelente lectura en párrafos largos.
- Monoespaciada para código, identificadores y métricas técnicas.
- Escala tipográfica compacta.

### 7.5 Densidad

La UI debe ofrecer dos modos:

- cómoda;
- compacta.

La versión generada inicialmente en Stitch debe usar densidad cómoda-profesional.

---

## 8. Estructura global de navegación

### 8.1 Navegación principal

```text
Sidebar
├── Nuevo chat
├── Conversaciones
├── Documentos
├── Colecciones
├── Agentes
├── Diagnóstico
└── Configuración
```

### 8.2 Barra superior

Debe contener:

- nombre o breadcrumb de la sección;
- estado del backend;
- modelo activo;
- selector de tema;
- menú de usuario o instancia local.

### 8.3 Panel contextual derecho

Debe ser reutilizable para:

- fuentes;
- detalles de recuperación;
- configuración de conversación;
- metadatos documentales;
- métricas de generación.

En móvil se convierte en drawer o modal inferior.

---

## 9. Mapa de pantallas que se generarán en Stitch

1. Shell principal de la aplicación.
2. Chat vacío.
3. Chat con conversación activa.
4. Panel de fuentes.
5. Inspector de recuperación.
6. Historial de conversaciones.
7. Lista de documentos.
8. Carga e indexación de documentos.
9. Detalle de documento.
10. Lista de colecciones.
11. Detalle de colección.
12. Lista de agentes.
13. Editor de agente.
14. Probador de agente.
15. Diagnóstico del sistema.
16. Configuración general.
17. Estados de error y desconexión.
18. Variantes responsive.

---

## 10. Pantalla principal: chat

### 10.1 Layout de escritorio

```text
┌───────────────┬──────────────────────────────┬──────────────────┐
│ Navegación    │ Conversación                 │ Fuentes/Detalles │
│               │                              │                  │
│ Nuevo chat    │ Encabezado                   │ Contexto usado   │
│ Historial     │ Mensajes                     │ Fragmentos       │
│ Documentos    │ Respuesta streaming          │ Metadatos        │
│ Colecciones   │                              │                  │
│ Agentes       │ Composer                     │                  │
└───────────────┴──────────────────────────────┴──────────────────┘
```

### 10.2 Encabezado de conversación

Debe mostrar:

- nombre editable de la conversación;
- agente activo;
- colecciones activas;
- modelo activo;
- botón para nueva conversación;
- menú de acciones;
- acceso a ajustes de la conversación.

### 10.3 Mensajes

Cada mensaje del usuario debe mostrar:

- contenido;
- hora opcional;
- acciones para copiar y editar/reintentar.

Cada respuesta del agente debe mostrar:

- nombre e icono del agente;
- respuesta renderizada en Markdown;
- citas inline;
- estado de generación;
- acciones: copiar, regenerar, valorar, exportar;
- resumen de fuentes;
- advertencia si la evidencia es insuficiente.

### 10.4 Composer

Debe incluir:

- textarea autosizable;
- envío con Enter y salto con Shift+Enter;
- botón enviar;
- botón detener durante streaming;
- selector rápido de agente;
- selector de colecciones;
- indicador de archivos/contexto adjunto;
- acceso a parámetros avanzados.

### 10.5 Estado vacío

Mostrar:

- saludo breve;
- agente activo;
- descripción de lo que puede hacer;
- cuatro sugerencias de preguntas;
- selector de colección;
- indicación de privacidad local.

---

## 11. Fuentes e inspector de recuperación

### 11.1 Panel de fuentes

Cada fuente se representa mediante una tarjeta compacta con:

- título del documento;
- tipo de archivo;
- página o sección;
- puntuación de relevancia;
- fragmento resumido;
- colección;
- acción para abrir detalle.

### 11.2 Vinculación de citas

Al pulsar una cita en la respuesta:

1. Se abre el panel de fuentes.
2. Se selecciona automáticamente la fuente relacionada.
3. Se resalta el fragmento utilizado.
4. Se muestra la página o ubicación.

### 11.3 Inspector avanzado

Debe mostrar:

- consulta original;
- consulta reformulada, si existe;
- resultados recuperados;
- score vectorial;
- score lexical;
- score final;
- ranking antes y después del reranker;
- tokens de contexto;
- tiempo de recuperación;
- filtros aplicados.

Este panel debe estar oculto por defecto para usuarios no técnicos.

---

## 12. Gestión de conversaciones

### 12.1 Lista

Debe permitir:

- buscar;
- ordenar por fecha;
- filtrar por agente;
- filtrar por colección;
- renombrar;
- duplicar;
- archivar;
- eliminar;
- exportar.

### 12.2 Presentación

En escritorio puede utilizar una tabla o lista densa. En móvil debe utilizar tarjetas.

Cada elemento muestra:

- título;
- extracto;
- agente;
- fecha;
- número de mensajes;
- colecciones relacionadas.

---

## 13. Gestión de documentos

### 13.1 Lista de documentos

Debe mostrar:

- nombre;
- formato;
- colección;
- tamaño;
- páginas;
- número de chunks;
- estado;
- fecha de carga;
- última indexación.

### 13.2 Estados documentales

- pendiente;
- extrayendo;
- fragmentando;
- generando embeddings;
- indexando;
- listo;
- error;
- desactualizado.

### 13.3 Carga de documentos

La pantalla debe permitir:

- arrastrar archivos;
- seleccionar múltiples archivos;
- asignar colección;
- añadir etiquetas;
- configurar idioma;
- elegir estrategia de chunking;
- iniciar indexación;
- ver progreso por archivo.

### 13.4 Detalle documental

Debe incluir:

- metadatos;
- vista previa;
- chunks;
- estado de embeddings;
- historial de indexación;
- errores;
- acciones de reindexar, mover o eliminar.

---

## 14. Gestión de colecciones

### 14.1 Lista

Cada colección muestra:

- nombre;
- descripción;
- número de documentos;
- número de chunks;
- modelo de embeddings;
- fecha de actualización;
- estado.

### 14.2 Detalle

Debe permitir:

- editar nombre y descripción;
- añadir o retirar documentos;
- configurar embeddings;
- definir estrategia de recuperación;
- reindexar;
- probar una consulta;
- ver métricas básicas.

---

## 15. Gestión de agentes

### 15.1 Lista de agentes

Cada tarjeta debe mostrar:

- nombre;
- propósito;
- colecciones permitidas;
- formato de respuesta;
- modelo;
- estado activo/inactivo;
- última modificación.

### 15.2 Editor de agente

El formulario debe organizarse en secciones:

#### Identidad

- nombre;
- descripción;
- icono;
- color identificativo.

#### Comportamiento

- system prompt;
- instrucciones de respuesta;
- tono;
- restricciones;
- mensaje cuando no hay evidencia.

#### Conocimiento

- colecciones disponibles;
- top K;
- filtros;
- reranker;
- score mínimo.

#### Generación

- modelo;
- temperatura;
- tokens máximos;
- formato de salida;
- esquema JSON opcional.

#### Capacidades

- memoria;
- herramientas;
- citas;
- razonamiento visible o resumido;
- acceso a funciones administrativas.

### 15.3 Probador de agente

Debe permitir enviar preguntas de prueba y ver simultáneamente:

- respuesta;
- fuentes;
- estructura JSON;
- errores de validación;
- tiempos;
- prompt final construido.

---

## 16. Diagnóstico del sistema

La pantalla de diagnóstico debe mostrar tarjetas y tablas con:

- estado del backend Go;
- estado de llama.cpp u Ollama;
- estado de Qdrant o pgvector;
- modelo cargado;
- modelo de embeddings;
- memoria utilizada;
- cola de indexación;
- latencia media;
- solicitudes recientes;
- errores recientes.

Debe evitar gráficas innecesarias. Priorizar métricas claras y estados operativos.

---

## 17. Configuración

Secciones:

- apariencia;
- conexión con backend;
- modelos;
- almacenamiento;
- privacidad;
- exportación;
- idioma;
- accesibilidad;
- configuración avanzada.

---

## 18. Estados especiales que Stitch debe diseñar

Para cada pantalla importante deben generarse al menos estos estados:

- cargando;
- vacío;
- contenido normal;
- error parcial;
- error total;
- desconectado;
- sin permisos;
- operación completada;
- modo responsive.

### 18.1 Backend desconectado

Mostrar:

- banner persistente;
- explicación breve;
- URL configurada;
- botón reintentar;
- acceso a configuración.

### 18.2 Sin fuentes suficientes

La respuesta debe mostrarse con una advertencia visible, sin aparentar certeza.

### 18.3 Generación detenida

Conservar el texto parcial y permitir continuar o regenerar.

---

## 19. Responsive design

### Escritorio

- sidebar persistente;
- panel central amplio;
- panel derecho contextual;
- alta densidad moderada.

### Tableta

- sidebar colapsable;
- panel contextual como drawer;
- composer siempre visible.

### Móvil

- navegación inferior o drawer lateral;
- una sola columna;
- fuentes en modal inferior;
- agentes y colecciones mediante selectores compactos;
- tablas transformadas en tarjetas.

Stitch debe generar al menos una variante móvil explícita de:

- chat;
- documentos;
- agentes.

---

## 20. Accesibilidad

La UI debe diseñarse para cumplir como mínimo:

- contraste AA;
- navegación por teclado;
- foco visible;
- etiquetas comprensibles;
- estados no comunicados solo por color;
- objetivos táctiles adecuados;
- soporte para lectores de pantalla;
- reducción de movimiento;
- tamaño de fuente legible.

---

## 21. Integración esperada con el backend Go

Stitch generará diseño y código de interfaz, pero la lógica real se conectará al backend Go.

### 21.1 Endpoints mínimos

```text
GET    /api/health
GET    /api/models

GET    /api/agents
POST   /api/agents
GET    /api/agents/:id
PUT    /api/agents/:id
DELETE /api/agents/:id
POST   /api/agents/:id/test

GET    /api/conversations
POST   /api/conversations
GET    /api/conversations/:id
PATCH  /api/conversations/:id
DELETE /api/conversations/:id

POST   /api/chat/query
POST   /api/chat/stream
POST   /api/chat/stop

GET    /api/collections
POST   /api/collections
GET    /api/collections/:id
PUT    /api/collections/:id
DELETE /api/collections/:id

GET    /api/documents
POST   /api/documents/upload
GET    /api/documents/:id
DELETE /api/documents/:id
POST   /api/documents/:id/reindex

GET    /api/jobs
GET    /api/diagnostics
```

### 21.2 Streaming

La UI debe estar diseñada para consumir SSE o streaming HTTP.

Eventos esperados:

```text
connected
retrieval_started
source_found
generation_started
token
validation_started
completed
cancelled
error
```

### 21.3 Modelo de respuesta

```json
{
  "conversation_id": "conv_123",
  "message_id": "msg_456",
  "agent": {
    "id": "researcher",
    "name": "Investigador"
  },
  "answer": "...",
  "sources": [
    {
      "id": "source_1",
      "document_id": "doc_1",
      "title": "Documento.pdf",
      "page": 24,
      "chunk_id": "chunk_82",
      "excerpt": "...",
      "score": 0.91
    }
  ],
  "metrics": {
    "retrieval_ms": 120,
    "generation_ms": 1850,
    "input_tokens": 2048,
    "output_tokens": 420
  }
}
```

---

## 22. Componentes reutilizables que Stitch debe mantener

- App shell.
- Sidebar item.
- Topbar.
- Status badge.
- Agent selector.
- Collection selector.
- Chat message.
- Citation chip.
- Source card.
- Document row.
- Collection card.
- Agent card.
- Metric card.
- Empty state.
- Error state.
- Loading skeleton.
- Confirmation modal.
- Drawer contextual.
- Form field.
- Advanced settings accordion.
- Streaming indicator.

Al generar nuevas pantallas, se debe pedir a Stitch que reutilice el mismo lenguaje visual y componentes en lugar de rediseñarlos.

---

## 23. Flujo de trabajo recomendado dentro de Stitch

### Fase 1 — Crear el sistema visual

Generar primero una pantalla shell con navegación, topbar y área de contenido vacía.

Objetivo:

- fijar colores;
- tipografía;
- espaciado;
- densidad;
- estilo de iconos;
- apariencia de controles.

### Fase 2 — Construir el chat

Generar:

1. estado vacío;
2. conversación activa;
3. respuesta en streaming;
4. respuesta con fuentes;
5. error y desconexión.

### Fase 3 — Construir administración documental

Generar:

1. documentos;
2. carga;
3. detalle;
4. colecciones.

### Fase 4 — Construir agentes

Generar:

1. lista;
2. editor;
3. probador.

### Fase 5 — Diagnóstico y configuración

Generar:

1. diagnóstico;
2. configuración;
3. responsive.

### Fase 6 — Refinamiento

Pedir a Stitch:

- reducir inconsistencias;
- reutilizar componentes;
- alinear espaciados;
- mejorar contraste;
- simplificar pantallas saturadas;
- producir variante móvil;
- mantener el mismo design system.

### Fase 7 — Exportación

- Exportar frontend.
- Revisar estructura generada.
- Sustituir datos mock por llamadas reales.
- Implementar el cliente SSE.
- Añadir validación y manejo de errores.
- Integrar autenticación si corresponde.

---

## 24. Prompts listos para Google Stitch

## 24.1 Prompt maestro del sistema visual

```text
Design a professional responsive web application for a local Retrieval-Augmented Generation platform. The backend is written in Go and the frontend will consume REST APIs and server-sent events.

Create a polished AI knowledge workspace, not a generic chatbot. Use a modern SaaS layout with a persistent left sidebar, a flexible central workspace, and an optional contextual right panel. The visual style should be sober, technical, trustworthy, and comfortable for long work sessions.

Use a dark professional theme with charcoal backgrounds, subtle borders, restrained elevation, clear typography, and one calm blue-green or indigo accent. Avoid neon colors, excessive gradients, glassmorphism, oversized cards, and decorative futuristic effects.

The navigation includes New Chat, Conversations, Documents, Collections, Agents, Diagnostics, and Settings. The top bar shows backend status, active model, theme control, and page context.

Establish a reusable design system for buttons, inputs, selectors, status badges, source cards, chat messages, tables, forms, drawers, modals, empty states, loading states, and error states. Prioritize accessibility, visible focus, strong contrast, keyboard usability, and responsive behavior.
```

## 24.2 Prompt para pantalla de chat vacía

```text
Using the established RAG application design system, create the empty state of the main chat screen.

Keep the left navigation visible. In the central workspace, show a compact welcome message, the currently selected agent, a collection selector, four useful suggested questions, and a privacy note explaining that the system runs locally.

At the bottom, create a professional autosizing chat composer with agent selector, collection selector, attachment/context indicator, advanced options button, and send button. The interface should feel ready for serious research work, with no oversized hero section and no marketing-style content.

The right contextual panel should be collapsed by default and available through a Sources button.
```

## 24.3 Prompt para conversación activa

```text
Using the same design system, create the active conversation view for the local RAG application.

Show a user question followed by a structured assistant response rendered as readable Markdown. Include inline citation chips such as [1] and [2], an agent identity, a subtle validation status, and actions for copy, regenerate, export, and feedback.

Open the right contextual panel and display source cards with document title, page, collection, relevance score, excerpt, and an action to inspect the source. Highlight the source connected to the selected citation.

The header should show the editable conversation title, active agent, active collections, model, and conversation settings. Keep the composer fixed at the bottom of the central workspace.
```

## 24.4 Prompt para streaming

```text
Create the streaming generation state for the active RAG chat screen using the existing layout and components.

The assistant response should be partially generated with a subtle live cursor. Show a compact progress sequence: Retrieving documents, Generating response, Validating citations. The completed steps should be visually distinct from the active step.

Replace the Send button with a prominent but restrained Stop button. Show source cards appearing progressively in the right panel as documents are retrieved. Avoid large loading animations.
```

## 24.5 Prompt para documentos

```text
Using the established design system, create the Documents management screen for the RAG application.

Include a searchable and filterable data table with columns for document name, type, collection, size, pages, chunks, indexing status, upload date, and last indexed date. Add bulk selection, upload, reindex, move, and delete actions.

Use clear compact status badges for Pending, Extracting, Chunking, Embedding, Indexing, Ready, Error, and Outdated. Add a side drawer for document details and a clear empty state.

The design should support high information density without feeling crowded.
```

## 24.6 Prompt para carga de documentos

```text
Create a document upload and indexing workflow for the professional RAG application.

Include drag-and-drop upload, multi-file selection, target collection, tags, language, chunking strategy, chunk size, overlap, and an advanced settings section. Show per-file validation and per-file indexing progress.

Design clear success, warning, and failure states. Keep technical parameters understandable by using helper text and progressive disclosure.
```

## 24.7 Prompt para agentes

```text
Using the same RAG design system, create the Agents management screen.

Show reusable agent cards with name, description, icon, active collections, model, response format, status, and last updated date. Include actions to create, edit, duplicate, test, deactivate, and delete an agent.

Add filters by model, collection, status, and response format. Keep the layout professional and suitable for both technical and non-technical users.
```

## 24.8 Prompt para editor de agente

```text
Create a professional multi-section Agent Editor for the local RAG application.

Organize the form into Identity, Behavior, Knowledge, Generation, Capabilities, and Validation. Include fields for name, description, icon, system prompt, response instructions, allowed collections, top K, score threshold, reranker, model, temperature, maximum tokens, output format, optional JSON schema, memory, citations, and tools.

Use a sticky action bar with Save, Save and Test, Cancel, and validation status. Include a live preview or testing panel on desktop. Use accordions or tabs to control complexity, but keep the system prompt editor prominent.
```

## 24.9 Prompt para diagnóstico

```text
Create the Diagnostics screen for the local RAG platform using the established design system.

Show operational status cards for the Go backend, vector database, LLM server, embedding model, storage, and indexing queue. Include current model, memory usage, average retrieval latency, average generation latency, active requests, queued jobs, and recent errors.

Use tables and compact metric cards rather than decorative charts. Clearly distinguish Healthy, Degraded, Offline, and Unknown states.
```

## 24.10 Prompt para versión móvil

```text
Create a mobile responsive version of the professional RAG chat interface while preserving the established visual system.

Use a single-column layout. Replace the persistent sidebar with a navigation drawer or compact bottom navigation. Keep the chat composer fixed and easy to use. Move sources into a bottom sheet or full-screen drawer. Agent and collection selectors should remain accessible without crowding the composer.

Ensure citation chips, Markdown responses, code blocks, and source cards remain readable on a narrow screen.
```

---

## 25. Prompts de refinamiento para Stitch

### Mantener consistencia

```text
Refine this screen so it strictly reuses the visual language, spacing, typography, navigation, form controls, cards, badges, and interaction patterns from the previously generated RAG screens. Do not introduce a new visual style.
```

### Reducir saturación

```text
Reduce visual clutter without removing functionality. Improve grouping, hierarchy, spacing, and progressive disclosure. Keep advanced technical controls available but secondary.
```

### Aumentar densidad profesional

```text
Make the interface more suitable for professional desktop work. Reduce oversized padding and cards, improve information density, and keep all text and controls readable.
```

### Mejorar accesibilidad

```text
Improve accessibility with stronger contrast, visible keyboard focus, larger touch targets where needed, clear labels, and states that do not rely only on color.
```

### Refinar responsive

```text
Adapt this exact screen to mobile and tablet while preserving component identity and hierarchy. Replace wide tables with cards, side panels with drawers, and keep primary actions immediately accessible.
```

---

## 26. Datos mock recomendados para generar las pantallas

### Agentes

```json
[
  {
    "name": "Investigador",
    "description": "Analiza documentos y responde con evidencias",
    "model": "Qwen Local",
    "collections": ["Patakíes", "Odu"],
    "status": "active"
  },
  {
    "name": "Profesor",
    "description": "Explica conceptos de forma progresiva",
    "model": "Qwen Local",
    "collections": ["Biblioteca general"],
    "status": "active"
  },
  {
    "name": "Editor",
    "description": "Revisa y estructura contenido editorial",
    "model": "Qwen Local",
    "collections": ["Manuscritos"],
    "status": "inactive"
  }
]
```

### Documentos

```json
[
  {
    "name": "Patakies_volumen_01.pdf",
    "type": "PDF",
    "collection": "Patakíes",
    "pages": 284,
    "chunks": 912,
    "status": "ready"
  },
  {
    "name": "Diccionario_yoruba.md",
    "type": "Markdown",
    "collection": "Lengua Yoruba",
    "pages": null,
    "chunks": 143,
    "status": "indexing"
  },
  {
    "name": "Notas_investigacion.docx",
    "type": "DOCX",
    "collection": "Investigación",
    "pages": 67,
    "chunks": 0,
    "status": "error"
  }
]
```

### Fuentes

```json
[
  {
    "title": "Patakies_volumen_01.pdf",
    "page": 24,
    "collection": "Patakíes",
    "score": 0.91,
    "excerpt": "Fragmento documental recuperado para respaldar la respuesta..."
  },
  {
    "title": "Diccionario_yoruba.md",
    "page": null,
    "collection": "Lengua Yoruba",
    "score": 0.84,
    "excerpt": "Definición relacionada con el concepto consultado..."
  }
]
```

---

## 27. Requisitos funcionales prioritarios

### P0 — Primera versión usable

- Shell y navegación.
- Chat vacío y activo.
- Selector de agente.
- Selector de colecciones.
- Streaming.
- Detener generación.
- Fuentes y citas.
- Historial básico.
- Documentos: lista y carga.
- Agentes: lista y edición básica.
- Estados de conexión.
- Responsive principal.

### P1 — Versión profesional

- Inspector de recuperación.
- Detalle documental.
- Colecciones completas.
- Probador de agentes.
- Diagnóstico.
- Exportación.
- Parámetros avanzados.
- Accesibilidad completa.

### P2 — Evolución

- Usuarios y permisos.
- Espacios de trabajo.
- Auditoría.
- Comparación de respuestas.
- Plantillas de agentes.
- Analítica avanzada.

---

## 28. Criterios de aceptación visual

La UI se considera visualmente aprobada cuando:

- todas las pantallas utilizan el mismo sistema visual;
- el chat es el foco principal;
- las fuentes son fáciles de localizar;
- los estados del sistema son evidentes;
- no hay gradientes, efectos ni decoraciones innecesarias;
- la densidad es adecuada para trabajo profesional;
- escritorio, tableta y móvil conservan la jerarquía;
- las pantallas administrativas no parecen un producto distinto;
- los formularios avanzados mantienen buena legibilidad;
- existe una versión clara de estados vacío, carga y error.

---

## 29. Criterios de aceptación funcional

- El frontend puede listar agentes y colecciones desde la API.
- El usuario puede iniciar una conversación.
- La respuesta se presenta mediante streaming.
- El usuario puede detener una generación.
- Las citas abren la fuente correspondiente.
- El panel de fuentes puede abrirse y cerrarse.
- El historial conserva conversaciones.
- Los documentos muestran progreso de indexación.
- Los agentes pueden crearse y editarse.
- Los errores del backend se comunican de forma accionable.
- La interfaz funciona mediante teclado.
- La versión móvil mantiene todas las acciones críticas.

---

## 30. Entregables esperados desde Stitch

1. Sistema visual base.
2. Pantalla de chat vacío.
3. Pantalla de conversación activa.
4. Estado de streaming.
5. Panel de fuentes.
6. Documentos.
7. Carga documental.
8. Colecciones.
9. Agentes.
10. Editor y probador de agente.
11. Diagnóstico.
12. Configuración.
13. Variantes móviles.
14. Código frontend exportado.
15. Opcionalmente, versión trasladada a Figma.

---

## 31. Revisión del código exportado

El código generado por Stitch debe tratarse como base visual, no como implementación final garantizada.

Antes de integrarlo se debe revisar:

- estructura de componentes;
- duplicación de estilos;
- accesibilidad;
- estado global;
- contratos TypeScript;
- manejo de errores;
- cliente HTTP;
- cliente SSE;
- rutas;
- seguridad;
- rendimiento;
- responsive real;
- eliminación de datos mock.

La lógica RAG, autenticación, persistencia, streaming y autorización permanecerán en el backend Go o en capas frontend implementadas manualmente.

---

## 32. Definición de terminado

La UI se considerará terminada cuando:

1. Las pantallas P0 hayan sido generadas y refinadas en Stitch.
2. El sistema visual sea consistente.
3. Existan variantes responsive críticas.
4. El código haya sido exportado.
5. Los datos mock hayan sido sustituidos por la API Go.
6. El streaming funcione correctamente.
7. Las fuentes estén vinculadas con las respuestas.
8. Los estados de error estén implementados.
9. La navegación por teclado sea funcional.
10. La aplicación pueda compilarse y desplegarse independientemente del backend.

---

## 33. Resultado esperado

El resultado final será una interfaz web profesional para un RAG local que pueda diseñarse de forma incremental en Google Stitch, exportarse a frontend y conectarse a un backend en Go.

El producto combinará:

- chat especializado;
- verificación documental;
- administración de conocimiento;
- configuración de agentes;
- diagnóstico técnico;
- experiencia responsive;
- arquitectura desacoplada.

El uso de Stitch debe acelerar la creación visual y la exploración de variantes, mientras que este PRD mantiene el control sobre la arquitectura, los flujos y los requisitos reales del producto.
