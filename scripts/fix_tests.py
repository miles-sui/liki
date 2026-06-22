#!/usr/bin/env python3
"""Comprehensive test fix for ChartBase simplification + API unification."""

import re, os, sys

TEST_DIRS = [
    'internal/engine/bazi',
    'internal/agent',
    'internal/http',
]

# ── Replacement rules ──

# Field renames (struct literals + field accesses)
FIELD_RENAMES = [
    (r'\bYear:', 'Nian:'),
    (r'\bMonth:', 'Yue:'),
    (r'\bDay:(?!\s*Gan|\s*Zhi)', 'Ri:'),
    (r'\bHour:', 'Shi:'),
]

ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []

# Removed from ChartBase (struct literal lines)
REMOVED_FIELDS = [
    r'\t+RiYuan:\s+[^,\n]+,\n',
    r'\t+FuYi:\s+FuYi\{[^}]+\},\n',
    r'\t+TiaoHou:\s+TiaoHou\{[^}]+\},\n',
    r'\t+WuxingCount:\s+[^,\n]+,\n',
    r'\t+DayMaster:\s+[^,\n]+,\n',
]

# Symbol renames
SYMBOL_RENAMES = [
    ('TenGod', 'ShiShen'),
    ('tenGod', 'shiShen'),
    ('tianwen.ComputeSolarTime', 'tianwen.ComputeTimeset'),
    ('"pillars"', '"zhu"'),
    ('current_pillar_index', 'current_zhu_index'),
    ('.Zhus', '.Zhu'),
    ('Zhus:', 'Zhu:'),
]

# JSON key renames (only in test anonymous structs, NOT in production)
JSON_KEY_RENAMES = [
    ('"day_stem"', '"ri_stem"'),
    ('"ri_yuan"', '"ri_gan"'),
]

# Old ComputeTimeset signature
def fix_compute_timeset(m):
    return (f'tianwen.GregorianToSolar('
            f'time.Date({m.group(1)}, time.Month({m.group(2)}), '
            f'{m.group(3)}, {m.group(4)}, {m.group(5)}, 0, 0, '
            f'time.FixedZone("", int({m.group(7)}*3600))), '
            f'{m.group(6)}, {m.group(7)})')

COMPUTE_TIMESET_RE = re.compile(
    r'tianwen\.ComputeTimeset\((\d+),\s*(\d+),\s*(\d+),\s*(\d+),\s*(\d+),\s*([\d.]+),\s*([\d.]+)\)'
)

# ComputeChart(st, g) where st is Timeset not SolarTime
COMPUTE_CHART_FIXES = [
    ('ComputeChart(st, g)', 'ComputeChart(st.Solar, g)'),
]

# ── Main ──

for test_dir in TEST_DIRS:
    dirpath = os.path.join('/home/sq/wsp/liki', test_dir)
    if not os.path.isdir(dirpath):
        continue
    for fname in os.listdir(dirpath):
        if not fname.endswith('_test.go'):
            continue
        fpath = os.path.join(dirpath, fname)
        
        with open(fpath) as f:
            content = f.read()
        
        original = content
        
        # Apply field renames
        for pat, repl in FIELD_RENAMES:
            content = re.sub(pat, repl, content)
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
ACCESS_RENAMES = []
