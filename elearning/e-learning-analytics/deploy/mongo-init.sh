until mongosh --host mongo1 --eval "db.adminCommand('ping')" &>/dev/null; do
  echo "Waiting for mongo1 to start..."
  sleep 2
done

# Initiate replica set
mongosh --host mongo1 <<EOF
rs.initiate({
  _id: "rs0",
  members: [
    {_id:0, host: "mongo1:27017"},
    {_id:1, host: "mongo2:27017"},
    {_id:2, host: "mongo3:27017"}
  ]
})
EOF

# Wait for primary election
while true; do
  PRIMARY=$(mongosh --host mongo1 --quiet --eval 'rs.isMaster().ismaster')
  if [ "$PRIMARY" == "true" ]; then
    break
  fi
  echo "Waiting for replica set primary..."
  sleep 3
done

# Create app user
mongosh --host mongo1 <<EOF
use elearning_auth
db.createUser({
  user: "elearn_user",
  pwd: "elearn_pass",
  roles: [ { role: "readWrite", db: "elearning_auth" } ]
})
EOF