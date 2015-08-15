#
#   Author: Rohith
#   Date: 2015-08-13 16:03:12 +0100 (Thu, 13 Aug 2015)
#
#  vim:ts=2:sw=2:et
#
FROM scratch:latest
ADD bin/rbd-lock-release /bin/rdb-manager
RUN chmod +x /bin/rdb-manager

ENRTYPOINT [ "/bin/rdb-manager" ]
