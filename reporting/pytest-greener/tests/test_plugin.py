import pytest


@pytest.fixture
def load_plugin(pytester):
    # plugin should be already installed,
    # so no "-p pytest_greener.plugin" when running pytest
    plugin_load_args = ""

    pytester.makefile(
        ".ini",
        pytest=f"""
        [pytest]
        addopts =
            {plugin_load_args}
    """,
    )


@pytest.mark.usefixtures("load_plugin")
def test_plugin_flat_structure(pytester):
    pytester.makepyfile(
        """
        def test_a():
            assert True

        def test_b():
            assert False
    """,
        test_file="""
        import pytest

        def test_c():
            assert True

        @pytest.mark.skip
        def test_d():
            assert False
    """,
    )

    # pytester.runpytest(
    #     "--greener",
    #     "--greener-address",
    #     # greener_address,
    # ).assert_outcomes(passed=2, failed=1, skipped=1)


@pytest.mark.parametrize(
    "arg",
    [
        "--greener",
        "--greener-address",
    ],
)
def test_plugin_not_loaded(pytester, arg):
    res = pytester.runpytest("-p", "no:greener", arg)

    res.stderr.fnmatch_lines(
        [
            "ERROR: usage:*",
            f"*error: unrecognized arguments: {arg}",
        ]
    )


@pytest.mark.usefixtures("load_plugin")
def test_plugin_loaded_no_args(pytester):
    pytester.makepyfile(
        """
        def test_a():
            assert True

        def test_b():
            assert True
    """
    )

    res = pytester.runpytest()

    res.assert_outcomes(passed=2)
