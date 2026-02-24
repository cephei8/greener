import pytest

def test_1_pass():
    assert True

def test_1_fail():
    assert False

@pytest.fixture
def fixture_with_error():
    raise ValueError("error in fixture")

def test_1_err(fixture_with_error):
    assert True

@pytest.mark.skip
def test_1_skip():
    assert False

class TestClass1:
    def test_1_pass(self):
        assert True