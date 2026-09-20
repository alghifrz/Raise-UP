# RAISE UP Frontend

Admin web application for RAISE UP (Phase 12A).

## Stack

- React 19
- Vite
- TypeScript (strict)
- Tailwind CSS
- React Router
- TanStack Query

## Setup

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

Open [http://localhost:5173](http://localhost:5173).

## Environment

| Variable | Description |
|---|---|
| `VITE_API_BASE_URL` | Backend API origin (example `http://localhost:8080`) |

Ensure the backend `CORS_ALLOWED_ORIGINS` includes `http://localhost:5173` (this is already the backend default).

## Scripts

```bash
npm run dev       # development server
npm run build     # typecheck + production build
npm run preview   # preview production build
npm run lint      # oxlint
```

## Phase 12A–12D scope

Implemented:

- Login (`/login`)
- Auth session bootstrap via `GET /api/v1/auth/me`
- Protected admin shell
- Dashboard (`/dashboard`)
- Residents CRUD UI (`/residents`)
- Complaints list/detail/status UI (`/complaints`, `/complaints/:id`)
- Announcements UI (`/announcements`, `/announcements/:id`) including publish
- Activities UI (`/activities`, `/activities/:id`) with reminder configuration
- Finance UI (`/finance`, `/finance/:id`) with summary cards
- Dues UI (`/dues`, `/dues/periods/:id`) with period summary and payment status

Placeholder routes (coming soon): gallery, village.
