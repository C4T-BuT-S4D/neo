#!/usr/bin/env python

import base64
import os
import time

while True:
    try:
        print(base64.b64encode(os.urandom(23), altchars=b'AB').decode().upper())
        time.sleep(1)
    except BaseException:  # noqa
        pass
