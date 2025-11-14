# servicio-gateway

Pequeño API Gateway en Go que reenruta operaciones a servicios de seguridad y perfiles, y publica eventos de eliminación.

Variables de entorno:
- SECURITY_URL (ej. http://user-service:8080/api/v1)
- PROFILE_URL  (ej. http://profile-service:8087/api)
- EVENT_BUS_URL (opcional, ej. http://notification-orchestrator:8080)
- PORT (por defecto 8080)

Endpoints:
- POST /auth/login
- POST /auth/register
- DELETE /users/{id}   -> reenvía a SECURITY_URL y publica evento user.deleted
- GET /users/{id}      -> une respuestas de SECURITY_URL /users/{id} y PROFILE_URL /profiles/{id}
- PUT /users/{id}      -> divide body en partes para security/profile y unifica respuestas

Ejecutar local:
SET SECURITY_URL=http://localhost:8080/api/v1
SET PROFILE_URL=http://localhost:8087/api
go run *.go

Construir imagen (Docker):
docker build -t servicio-gateway:local .
