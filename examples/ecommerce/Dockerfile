# Build from the repository root so the local Forge module is available.
FROM node:26.8.1-alpine AS admin
WORKDIR /src/forge/admin/ui/web
COPY forge/admin/ui/web/package*.json ./
RUN npm ci
COPY forge/admin/ui/web/ ./
RUN npm run build

FROM golang:1.27.1-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY forge/go.mod forge/go.sum ./forge/
COPY examples/ecommerce/go.mod examples/ecommerce/go.sum ./examples/ecommerce/
WORKDIR /src/examples/ecommerce
RUN go mod download
COPY forge/ /src/forge/
COPY examples/ecommerce/ ./
COPY --from=admin /src/forge/admin/ui/dist /src/forge/admin/ui/dist
RUN CGO_ENABLED=1 go build -tags embed -o /out/ecommerce .

FROM alpine:latest
RUN apk add --no-cache ca-certificates && \
    apk upgrade --no-cache libssl3 libcrypto3 && \
    addgroup -g 1000 app && adduser -D -u 1000 -G app app
WORKDIR /home/app
COPY --from=builder --chown=app:app /out/ecommerce ./
COPY --from=builder --chown=app:app /src/examples/ecommerce/config ./config
COPY --from=builder --chown=app:app /src/examples/ecommerce/templates ./templates
COPY --from=builder --chown=app:app /src/examples/ecommerce/static ./static
RUN chown app:app /home/app
USER app
ENV FORGE_SERVER_HOST=0.0.0.0 FORGE_SERVER_PORT=8020
EXPOSE 8020
HEALTHCHECK --interval=30s --timeout=3s --start-period=30s --retries=3 \
  CMD wget --quiet --tries=1 --spider http://localhost:8020/health || exit 1
CMD ["./ecommerce"]
