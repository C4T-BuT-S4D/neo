#!/usr/bin/env python

import base64
import os

for i in range(1000):
    print('A' * 2 * 1024 * 1024)
    print(base64.b64encode(os.urandom(23), altchars=b'AB').decode().upper())
