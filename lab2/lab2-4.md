# Lab II.4 — RESTful APIs & gRPC: InstaSafe Device Enrolment API

**Course:** RESTful APIs & gRPC
**Week:** 5-6
**Role:** Developer

## Objective

Design and document a complete OpenAPI 3.0 specification for InstaSafe's Device Enrolment API, covering all CRUD operations, authentication, request/response schemas, and error codes.

---

## What I built

Using Swagger Editor, I wrote an OpenAPI 3.0 spec (`openapi.yaml`) for a Device Enrolment API with:

- **Bearer/JWT authentication** applied globally via a `BearerAuth` security scheme
- **Three schemas**: `Device`, `DeviceEnrolmentRequest`, and `ErrorResponse`, all reused across endpoints with `$ref`
- **Five endpoints** covering the full device lifecycle:
  - `POST /devices` — enrol a new device
  - `GET /devices` — list all devices (with `status`, `page`, `limit` query params)
  - `GET /devices/{deviceId}` — get a single device
  - `PATCH /devices/{deviceId}` — update device status
  - `DELETE /devices/{deviceId}` — unenrol a device
- **Realistic examples** on every request/response (UUIDs, hostnames like `DESKTOP-INSTASAFE-042`, base64-encoded certificate strings, ISO timestamps)
- **Error handling** — every endpoint documents `401 Unauthorized` (since auth is global), and every endpoint with a `{deviceId}` path param documents `404 Not Found`

---

## Screenshot 1: Swagger Editor — zero errors, all endpoints rendered

*(Insert `swagger-screenshot.png` here)*

`![Swagger Editor screenshot](./swagger-screenshot.png)`

This shows all 5 endpoints rendered correctly in Swagger UI's preview panel, color-coded by HTTP verb, with the lock icon confirming auth is applied to each. No errors in the validation panel.

---

## Screenshot 2: Postman import — 5 endpoints visible

*(Insert `postman-import.png` here)*

`![Postman import screenshot](./postman-import.png)`

Importing the exported `openapi.yaml` into Postman correctly generated a collection with all 5 endpoints, organized into a `devices` folder and a `{deviceId}` subfolder, along with the saved response examples (e.g. "Device successfully enrolled," "Unauthorized," "Device already enrolled").

---

## Findings & Design Decisions

### 1. Why PATCH instead of PUT for updating device status

I used PATCH for the `/devices/{deviceId}` status update endpoint instead of PUT, because PUT is meant to replace the entire resource, and that's not what's actually happening here. When we update a device's status (say, moving it from `pending` to `blocked`), we're only changing one field, not resending the whole device object. If I used PUT, the client would technically need to send back the full Device resource (id, hostname, os, enrolledAt, etc.) just to change one value, which doesn't make sense since fields like `id` and `enrolledAt` are server-generated and shouldn't be something the client resends anyway. PATCH is built exactly for partial updates like this, so the request body only needs `{ "status": "blocked" }`, which is simpler for the client and matches what's actually happening on the backend.

### 2. Why a dedicated inline request body instead of reusing DeviceEnrolmentRequest or the full Device schema

For the PATCH endpoint, I didn't reuse `DeviceEnrolmentRequest` or the full `Device` schema, and instead wrote a small dedicated inline schema with just the `status` field. `DeviceEnrolmentRequest` didn't fit because it's meant for enrolling a brand new device (hostname, os, deviceCertificate), and none of those fields make sense when you're just updating an existing device's status. Using the full `Device` schema also didn't make sense, because that schema includes server-generated fields like `id` and `enrolledAt`, and I didn't want to give the impression that a client could (or should) send those back. A small, purpose-built schema keeps the PATCH endpoint's contract narrow and makes it obvious to anyone reading the spec that this endpoint does exactly one thing: change the status, nothing else.

---

## What I learned

Building this spec end-to-end (info → auth → schemas → endpoints → examples → validation → Postman import) made it clear how much `$ref` matters for keeping a spec maintainable — writing `Device` and `ErrorResponse` once and referencing them across 5 endpoints instead of repeating the JSON shape every time. It also made me think harder about REST semantics I'd normally take for granted, like why PATCH exists as a separate verb from PUT, and why marking `required` fields on the request schema (versus the response schema) actually matters, since server-generated fields like `id` should never be something a client is expected to send.

## Submission Checklist

- [x] `openapi.yaml` — exported from Swagger Editor
- [x] `swagger-screenshot.png` — zero errors visible
- [x] `postman-import.png` — 5 endpoints visible
- [x] Design decisions explained (2, ~100 words each)
