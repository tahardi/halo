# API

The backend API is defined by JSON test data in `testdata/`. It is not app data: the files are examples used only by
Go and Jest tests, so the two sides cannot drift.

**Change a file in `testdata/` only together with both the Go and TypeScript code in the same PR.**

Every request, including `/health`, sends `Authorization: Bearer <key>`. Uploads are limited to 50 MB.

Every request can also fail with `401` (`testdata/error-unauthorized.json`) or `500` (`internal`).

## Endpoints

### GET /health

- Body: none
- Success: `200` with `testdata/health-response.json`

### PUT /attacks/{id}

Creates or replaces one attack.

- Body: an `Attack` matching `testdata/attack.json`. The server ignores `source`, `created_at`, and `updated_at`.
- Success: `200` with the stored `Attack` (`testdata/attack.json`)
- Failure: `400` (`testdata/error-bad-request.json`) for malformed JSON, an invalid id, severity outside 0 to 10, an
  unknown symptom, an unknown medication id in a dose, or `ended_at` before `started_at`

### GET /attacks?before=\<RFC 3339\>&limit=\<1-200\>

Lists attacks that started before `before`, newest first.

- Body: none
- Query: `before` defaults to now; `limit` defaults to 50
- Success: `200` with an `AttackPage` (`testdata/attack-page.json`)
- Failure: `400` (`testdata/error-bad-request.json`) for a bad `before` or `limit`

### GET /attacks/{id}

- Body: none
- Success: `200` with an `AttackDetail` (`testdata/attack-detail.json`). `weather` is `null` until weather is known.
- Failure: `404` (`testdata/error-not-found.json`) for an unknown or deleted attack

### DELETE /attacks/{id}

Soft-deletes one attack.

- Body: none
- Success: `204` with no body
- Failure: `404` (`testdata/error-not-found.json`)

### GET /medications

- Body: none
- Success: `200` with `{"medications":[Medication]}` (`testdata/medication-list.json`), archived included, sorted by
  name

### PUT /medications/{id}

Creates or replaces one medication.

- Body: a `Medication` matching an entry in `testdata/medication-list.json`
- Success: `200` with the stored `Medication`
- Failure: `400` (`testdata/error-bad-request.json`)

### POST /attacks/{id}/recordings

Uploads one voice recording for an attack, then transcribes it and proposes changes.

- Headers: `Content-Type: multipart/form-data`
- Multipart fields: `id`, `recorded_at`, `audio`
- Success: `200` with a `Recording` (`testdata/recording-extracted.json`). The same recording id again returns the
  stored `Recording` unchanged.
- Failure:
  - `400` (`testdata/error-bad-request.json`)
  - `404` (`testdata/error-not-found.json`) for an unknown attack
  - `413` (`payload_too_large`) for an upload over 50 MB
  - `422` (`testdata/error-no-speech.json`) for an empty transcript
  - `502` (`testdata/error-upstream-failed.json`) for an ffmpeg, whisper, or model failure. The backend keeps the
    audio, and the recording reads back as `testdata/recording-failed.json`.

### POST /recordings/{id}/retry

Runs transcription and extraction again for a recording.

- Body: none
- Success: `200` with a `Recording` (`testdata/recording-extracted.json`)
- Failure: `404` (`testdata/error-not-found.json`), `422` (`testdata/error-no-speech.json`), or `502`
  (`testdata/error-upstream-failed.json`)

### POST /recordings/{id}/confirm

Marks a recording's proposal as applied.

- Body: none
- Success: `200` with a `Recording` whose `status` is `confirmed`
- Failure: `404` (`testdata/error-not-found.json`)

## Error codes

Every non-2xx response has the body `{"error":{"code":"<code>","message":"<text>"}}`.

| Status | Code                | When                                                                                                                                   | Fixture                               |
| ------ | ------------------- | -------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| 400    | `bad_request`       | malformed JSON, invalid id, severity outside 0 to 10, unknown symptom, unknown medication id in a dose, `ended_at` before `started_at` | `testdata/error-bad-request.json`     |
| 401    | `unauthorized`      | missing or wrong key                                                                                                                   | `testdata/error-unauthorized.json`    |
| 404    | `not_found`         | unknown attack, recording, or medication id                                                                                            | `testdata/error-not-found.json`       |
| 413    | `payload_too_large` | upload over 50 MB                                                                                                                      | none                                  |
| 422    | `no_speech`         | empty transcript                                                                                                                       | `testdata/error-no-speech.json`       |
| 502    | `upstream_failed`   | ffmpeg, whisper, or model failure                                                                                                      | `testdata/error-upstream-failed.json` |
| 500    | `internal`          | unexpected store or file error                                                                                                         | none                                  |

## Rules

- Times are RFC 3339 strings. The backend returns UTC.
- Ids are 1 to 64 characters of `[A-Za-z0-9-]`.
- Lists encode as `[]`, never `null`.
- `PUT /attacks/{id}` replaces the attack's doses with the body's doses.
- Uploading the same recording id twice returns the stored recording.
