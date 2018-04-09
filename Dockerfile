FROM centurylink/ca-certs

LABEL maintainer="Geofrey Ernest"

ADD bq /usr/local/bin/vectypresent

ENTRYPOINT ["/usr/local/bin/vectypresent"]

