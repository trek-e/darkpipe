# Rspamd web UI controller configuration
# Access at http://localhost:11334

# Password for read-only access (view stats and reports)
# Generated with: echo 'yourPassword' | rspamadm pw
# Default password: changeThisPassword123
# IMPORTANT: Change this password before production deployment
password = "$2$ky8w8dzmtkke5meupo3ux8tptjgo8cay$wi1m1ykjo6i4eqtqctf6k1sm3u5b8e1imnzjnmhdbm1hn37nciby";

# Password for read-write access (modify configuration)
# Using same hash as password for simplicity - change in production
enable_password = "$2$ky8w8dzmtkke5meupo3ux8tptjgo8cay$wi1m1ykjo6i4eqtqctf6k1sm3u5b8e1imnzjnmhdbm1hn37nciby";

# Bind to all interfaces on port 11334 (accessible within Docker network and via port mapping)
bind_socket = "*:11334";
