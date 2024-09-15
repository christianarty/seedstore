FROM scratch
COPY seedstore /seedstore
ENTRYPOINT ["/seedstore"]