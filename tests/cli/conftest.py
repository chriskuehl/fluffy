import os
import sys
from unittest import mock

import pytest


@pytest.yield_fixture
def cli_on_path():
    with mock.patch.dict(
            os.environ,
            {
                'PATH': '{}:{}'.format(
                    os.path.dirname(sys.executable), os.environ['PATH'],
                ),
            },
    ):
        yield
