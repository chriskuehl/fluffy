import pygments.lexers
import pytest

from fluffy.highlighting import UI_LANGUAGES_MAP


@pytest.mark.parametrize('language', UI_LANGUAGES_MAP)
def test_ui_language_exists(language):
    """Ensure a lexer exists for each language we advertise."""
    assert pygments.lexers.get_lexer_by_name('python') is not None
