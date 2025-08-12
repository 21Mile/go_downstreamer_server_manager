# --- runtime only ---
FROM alpine:3.20

# 如程序需要访问 https 资源，建议安装 CA 证书；同时设置时区（可选）
RUN apk --no-cache add ca-certificates tzdata \
    && update-ca-certificates

# 可选：默认时区（按需开启）
# ENV TZ=Asia/Shanghai

WORKDIR /app

# 从宿主机拷贝交叉编译好的二进制
# 确保 build.sh 已将 ./bin/downstream-server-manager 生成好
COPY ./bin/downstream-server-manager /app/downstream-server-manager

# （可选）拷贝配置目录（如果仓库中存在）
# 没有该目录就删除这行，运行时用 -v 挂载也更灵活
# COPY ./conf/prod/ /app/conf/prod/

# 入口
ENTRYPOINT ["/app/downstream-server-manager"]

# （可选）默认参数，按你的程序实际需要启用
# CMD ["-configPath=/app/conf/prod/", "-endpoint=server"]
