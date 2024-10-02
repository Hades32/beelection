FROM node:18-alpine AS node-build
# RUN apk add --no-cache libc6-compat
WORKDIR /app
COPY ./frontend/package* ./
RUN --mount=type=cache,target=/root/.npm \
    npm ci
COPY ./frontend/ ./
RUN --mount=type=cache,target=/root/.npm \
    npm run build

FROM golang:1.23.2 as go-build
WORKDIR /app
COPY ./backend/go.* /app/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=ssh \
    go mod download
COPY ./backend/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build $GO_ARGS -o /app/tangia

FROM gcr.io/distroless/base
WORKDIR /app
ENTRYPOINT ["/app/tangia"]
ENV PORT=8080
COPY --from=go-build /app/tangia /app/
COPY --from=node-build /app/build /frontend/build
