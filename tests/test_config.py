from pathlib import Path

from marge.config import load_config


def test_load_config_defaults_prune_css_false(tmp_path: Path) -> None:
    config_path = tmp_path / "config.toml"
    config_path.write_text('title = "Site"\nbase_url = "https://example.com"\n')
    config = load_config(config_path)
    assert config.prune_css is False
    assert config.prune_css_safelist == frozenset()


def test_load_config_reads_prune_css_settings(tmp_path: Path) -> None:
    config_path = tmp_path / "config.toml"
    config_path.write_text(
        'title = "Site"\n'
        'base_url = "https://example.com"\n'
        "prune_css = true\n"
        'prune_css_safelist = ["is-open", "active"]\n'
    )
    config = load_config(config_path)
    assert config.prune_css is True
    assert config.prune_css_safelist == frozenset({"is-open", "active"})
