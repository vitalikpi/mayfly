# ---- build stage ----
FROM golang:1.27 AS build
WORKDIR /src
COPY . .
# Static binary (no libc), stripped of debug info. All HTML/JS/CSS is embedded.
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /mayfly .

# ---- run stage ----
# scratch = empty image: nothing but the binary.
FROM scratch
COPY --from=build /mayfly /mayfly
# Run as a non-root uid (nobody). Numeric id works without /etc/passwd.
USER 65534:65534
EXPOSE 8080
ENTRYPOINT ["/mayfly"]
