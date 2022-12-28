.. |form| image:: ./image/form.png
   :alt: Input Form

.. |output| image:: ./image/output.png
   :alt: Sample Output

.. |architecture| image:: ./image/architecture.png
   :alt: General Architecture


+----------------+
| |architecture| |
+----------------+

Links
-----

* `Docker Docs <https://docs.docker.com/>`_
* `Docker Compose getting started <https://docs.docker.com/compose/gettingstarted/>`_
* `Docker Compose file reference <https://docs.docker.com/compose/compose-file/>`_
* `Apache HBase Reference Guide <http://hbase.apache.org/book.html>`_
* `ZooKeeper Documentation <http://zookeeper.apache.org/doc/trunk/>`_
* `Go Documentation <https://golang.org/doc/>`_
* `Pro Git <https://git-scm.com/book/en/v2>`_

Components
----------

Nginx
~~~~~

Nginx is a web server that delivers static content in our architecture.
Static content comprises the landing page (index.html), JavaScript, css and font files located in ``nginx/www``.

HBase
~~~~~

We use HBase, the open source implementation of Bigtable, as database.
``hbase/hbase_init.txt`` creates the ``se2`` namespace and a ``library`` table with two column families: ``document`` and ``metadata``.

   * `HBase REST documentation <http://hbase.apache.org/book.html#_rest>`_
   * The client port for REST is 8080
   * Use Curl to explore the API
      * ``curl -vi -X PUT -H "Content-Type: application/json" -d '<json row description>' "localhost:8080/se2:library/fakerow"``
   
ZooKeeper
~~~~~~~~~

**The HBase image above already contains a ZooKeeper instance.**

Grproxy
~~~~~~~

A reverse proxy that forwards every request to nginx, except those with a "library" prefix in the path (e.g., ``http://host/library``).
We discover running gserve instances with the help of the ZooKeeper service and forward ``library`` requests in circular order among those instances (Round Robin).

Gserve
~~~~~~

Gserve serves two purposes.
Firstly, it receives ``POST`` requests from the client (via grproxy) and adds or alters rows in HBase.
And secondly, it replies to ``GET`` requests with an HTML page displaying the contents of the whole document library.
It only receives requests from grproxy after it subscribed to ZooKeeper, and automatically unsubscribes from ZooKeeper if it shuts down or crashes.


***NOTE***: In case HBase build is broken

Unfortunately, HBase does not maintain a stable URL to the latest version of the software and we have to periodically migrate to newer versions.

Try changing the variable HBASE_VERSION in hbase/Dockerfile to a more recent version. You can find a list of available versions here: http://apache.lauf-forum.at/hbase/stable/
  

To start use the command ``docker-compose up``
