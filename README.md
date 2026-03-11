# GlobalTask Bank - Credit Application System (Fintech Core)

Este sistema es una solución de backend y frontend de alto rendimiento diseñada para gestionar solicitudes de crédito a escala global. Construido sobre una arquitectura limpia (**Clean Architecture**) y diseño orientado al dominio (**DDD**), el sistema soporta flujos de trabajo asíncronos distribuidos, procesamiento concurrente y actualizaciones en tiempo real mediante **WebSockets**.

---

## 🚀 1. Instrucciones de Instalación y Ejecución

### Requisitos Previos
*   **Docker** y **Docker Compose** instalados.
*   **Git** para clonar el repositorio.

### Levantamiento Rápido (Docker Compose)
1.  Clona el repositorio.
2.  Copia el archivo de ejemplo de variables de entorno (si existe) o usa el `.env` base.
3.  Ejecuta el siguiente comando en la raíz del proyecto:
    ```bash
    docker-compose up --build -d
    ```
4.  **Acceso:**
    *   **Frontend (Dashboard):** [http://localhost](http://localhost)
    *   **API Health:** [http://localhost/health](http://localhost/health)
    *   **WebSocket Endpoint:** `ws://localhost/api/v1/ws`
    *   **Mock Bank:** [http://localhost:8081](http://localhost:8081) (Solo para el backend).

---

## 🧠 2. Supuestos y Decisiones Técnicas

### Supuestos del Negocio
*   El usuario se autentica a través del servicio de Supabase (JWT). El backend valida criptográficamente la firma del JWT.
*   Un banco puede tardar varios segundos en responder, por lo que la comunicación debe ser asíncrona mediante webhooks.
*   Diferentes países tienen diferentes umbrales de riesgo y formatos de documentos de identidad (CURP en MX, NIF en PT).

### Decisiones Técnicas (Arquitectura)
*   **Go (Golang):** Elegido por su manejo nativo de concurrencia (**Goroutines**), bajo consumo de memoria y tipado fuerte, ideal para procesar miles de créditos en paralelo.
*   **Clean Architecture:** El dominio está aislado de la infraestructura. Si mañana decidimos cambiar **Postgres (Supabase)** por **MongoDB** o **Redis**, la lógica de negocio no se ve afectada.
*   **Monorepo:** Facilita el despliegue coordinado de micro-servicios relacionados (`api`, `worker`, `mockbank`).
*   **Unit of Work (UoW):** Implementado para garantizar que la creación de un crédito y su encolamiento en el outbox sean una operación **atómica**.

---

## 📊 3. Modelo de Datos

Usamos un esquema relacional optimizado en **PostgreSQL**:

*   **`users`**: Gestionados por Supabase Auth, integrados mediante ID de UUID.
*   **`countries`**: Configuración estática para Portugal (PT) y México (MX).
*   **`loan_applications`**: Tabla central con estados (`PENDING`, `AWAITING_BANK_DATA`, `APPROVED`, etc.).
*   **`event_outbox`**: Nuestra cola persistente. Almacena cada paso que el sistema debe dar de forma asíncrona.
*   **`webhook_logs`**: Auditoría completa de las respuestas recibidas de proveedores externos.

---

## 🔒 4. Consideraciones de Seguridad
1.  **Validación JWT (Supabase):** El backend no confía en el frontend. Cada petición a la API debe incluir un JWT válido firmado por el secreto del proyecto de Supabase.
2.  **Rate Limiting (Nginx):** Implementado a nivel de proxy para prevenir ataques de denegación de servicio (DoS) y fuerza bruta en los webhooks.
3.  **RLS (Row Level Security):** Configurado en Postgres para asegurar que los usuarios solo puedan ver sus propios créditos, incluso si hubiera una brecha en la capa de aplicación.
4.  **Least Privilege:** El backend usa usuarios con permisos restringidos a nivel de DB.

---

## 📈 5. Escalabilidad y Grandes Volúmenes de Datos

*   **Escalabilidad Horizontal:** Los `Workers` pueden replicarse indefinidamente. Gracias al patrón `FOR UPDATE SKIP LOCKED` de Postgres, múltiples workers pueden leer de la misma tabla de eventos sin bloquearse ni duplicar tareas.
*   **Particionamiento:** Para volúmenes masivos (>100M de registros), la tabla `loan_applications` puede particionarse por `country_id` o `created_at`.
*   **Concurrency Strategy:** El sistema utiliza el modelo **CSP (Communicating Sequential Processes)** de Go, delegando tareas pesadas a goroutines ligeras en lugar de hilos de sistema pesado.

---

## 🔄 6. Estrategia de Concurrencia, Colas y Webhooks

### Colas (Event Outbox)
En lugar de depender de un broker externo (como RabbitMQ) para la prueba técnica, usamos **Event Outbox en Postgres**. Esto garantiza que el evento se guarde **solo si** la base de datos confirma la transacción del crédito. Esto elimina el riesgo de "créditos perdidos".

### Webhooks
El sistema es **Event-Driven**. Cuando el banco (Mock) termina su procesamiento, golpea nuestro endpoint de webhook. Esto dispara un nuevo evento en el Outbox que el Worker toma para realizar la decisión final de riesgo.

### WebSockets (Real-time)
Implementamos un **WS Hub nativo**. Los cambios de estado generados por el Worker o por Webhooks se transmiten instantáneamente al frontend. El usuario ve cómo su crédito se aprueba en tiempo real sin necesidad de refrescar la página.

---

## ☸️ 7. Despliegue en Kubernetes

Los manifiestos se encuentran en la carpeta `/infra/k8s/`. Incluyen:
*   **Deployments:** Réplicas configurables para la API y los Workers.
*   **Services:** LoadBalancers internos para comunicación entre servicios.
*   **ConfigMaps/Secrets:** Gestión de variables de entorno sensibles.
*   **Ingress:** Reglas de ruteo para el frontend y la API.

---

**Desarrollado con ❤️ para GlobalTask Bank.**
