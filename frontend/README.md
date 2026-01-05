booking/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── auth/
│   │   ├── booking/
│   │   ├── user/
│   │   ├── payment/
│   │   ├── notification/
│   │   └── middleware/
│   ├── migrations/
│   │   └── 001_create_tables.sql
│   ├── go.mod
│   ├── Dockerfile
│   └── .env.example
│
├── frontend/
│   ├── app/
│   │   ├── page.tsx
│   │   ├── booking/page.tsx
│   │   ├── dashboard/page.tsx
│   │   └── admin/page.tsx
│   ├── components/
│   ├── lib/api.ts
│   ├── tailwind.config.js
│   └── next.config.js
│
├── docker-compose.yml
└── README.md
