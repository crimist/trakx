echo "Make sure go_expvar = yes in python.d.conf"

cp expvar.conf /etc/netdata/python.d/go_expvar.conf
cp alarm.conf /etc/netdata/health.d/trakx.conf

service netdata restart
