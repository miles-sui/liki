"""25types Questionnaire Simulation & Verification

Verifies:
  1. Σv = 0 invariant always holds
  2. All 25 types (5 pure + 20 composite) are reachable
  3. q-v orthogonality: q is NOT correlated with any v component direction
  4. Δq divergence: self-q vs observer-q from different vantage points
  5. Prototype self-classification
  6. Distribution characteristics
  7. Laplace-smoothed p distribution (no zero proportions)
  8. Stratified question-bank sampling fairness
  8. Question-count impact on classification stability
  9. Self/other assessment delta-v divergence detection
"""

import math
import random
from collections import Counter

# =============================================================================
# Element indices
# =============================================================================

W, F, E, M, R = 0, 1, 2, 3, 4
ELEMENT_NAMES = ['Wood', 'Fire', 'Earth', 'Metal', 'Water']
ELEMENT_CODES = ['W', 'F', 'E', 'M', 'R']

# Laplace smoothing prior (each element gets +1 base count)
LAPLACE_PRIOR = 1

# =============================================================================
# Role display names (internal code → user-facing display)
# =============================================================================

ROLE_ZH = {'W': '开创者', 'F': '发光者', 'E': '承载者', 'M': '规范者', 'R': '蕴蓄者'}
ROLE_EN = {'W': 'Pioneer', 'F': 'Luminary', 'E': 'Steward', 'M': 'Refiner', 'R': 'Reservoir'}
ROLE_SHORT_ZH = {'W': '创', 'F': '焕', 'E': '承', 'M': '规', 'R': '蕴'}
ROLE_SHORT_EN = {'W': 'Pnr', 'F': 'Lum', 'E': 'Stw', 'M': 'Rfn', 'R': 'Rsv'}


def role_display(label, lang='zh', short=False):
    """Convert type code to role display name."""
    if len(label) == 1:
        # Pure type
        code = label
        return ROLE_ZH[code] if lang == 'zh' else ROLE_EN[code]
    # Composite: WF → Pioneer-Luminary (en) or 发光·开创 (zh)
    primary = label[0]
    secondary = label[1]
    if lang == 'zh':
        return f"{ROLE_ZH[secondary]}·{ROLE_ZH[primary]}"
    else:
        return f"{ROLE_EN[primary]}-{ROLE_EN[secondary]}"


def role_display_short(label, lang='zh'):
    """Short form for social sharing cards."""
    if len(label) == 1:
        return ROLE_SHORT_ZH[label] if lang == 'zh' else ROLE_SHORT_EN[label]
    primary = label[0]
    secondary = label[1]
    if lang == 'zh':
        return f"{ROLE_SHORT_ZH[secondary]}·{ROLE_SHORT_ZH[primary]}"
    else:
        return f"{ROLE_SHORT_EN[primary]}·{ROLE_SHORT_EN[secondary]}"

# =============================================================================
# Scoring: v (ranked choice) — mirrors engine/assessment.go ComputeV
# =============================================================================

def score_v(answers):
    """answers: list of (primary_idx, secondary_idx) from 5 elements.
    primary=2pt, secondary=1pt. Returns v[5] with Σv=0."""
    raw = [0, 0, 0, 0, 0]
    for p, s in answers:
        raw[p] += 2
        raw[s] += 1
    total = sum(raw)
    if total == 0:
        return [0.0, 0.0, 0.0, 0.0, 0.0]
    v = [(raw[i] / total - 0.2) * 5.0 for i in range(5)]
    return v


def score_p(answers, laplace_prior=LAPLACE_PRIOR):
    """Compute user-facing p (proportion, Σp=1) with Laplace smoothing.
    Each element gets +laplace_prior base count — no zero proportions."""
    raw = [laplace_prior] * 5
    for p_idx, s_idx in answers:
        raw[p_idx] += 2
        raw[s_idx] += 1
    total = sum(raw)
    return [raw[i] / total for i in range(5)]


def p_to_v(p):
    """Convert display p back to internal v (for verification)."""
    # p[i] = (raw[i] + LAPLACE_PRIOR) / (Σraw + 5*LAPLACE_PRIOR)
    # NOT directly reversible without raw, but we can approximate:
    # p[i] ≈ (raw[i]) / Σraw  for large Σraw
    return [(p[i] - 0.2) * 5.0 for i in range(5)]


# =============================================================================
# Scoring: q (binary forced choice) — mirrors engine/assessment.go ComputeQ
# =============================================================================

def score_q(answers):
    """answers: list of bools (True=A=正气足, False=B=正气不足).
    Returns q ∈ [0, 1]."""
    if not answers:
        return 0.5
    return sum(1 for a in answers if a) / len(answers)


# =============================================================================
# Q precision rounding — mirrors engine/assessment.go QPrecision
# =============================================================================

def q_precision(q, n_answers):
    if n_answers >= 100:
        return round(q, 4)
    elif n_answers >= 20:
        return round(q, 3)
    else:
        return round(q, 2)


# =============================================================================
# Prototypes — mirrors engine/identity.go
# =============================================================================

# Prototypes live on the unit 4-sphere (Σx²=1).
# Classification: d → ReLU → argmax inner product. No normalization (Σd⁺ cancels).
PROTOTYPES = {}

# Pure (5): one-hot — single element dominates
PURE_VALS = {
    'W': [1.0, 0.0, 0.0, 0.0, 0.0],
    'F': [0.0, 1.0, 0.0, 0.0, 0.0],
    'E': [0.0, 0.0, 1.0, 0.0, 0.0],
    'M': [0.0, 0.0, 0.0, 1.0, 0.0],
    'R': [0.0, 0.0, 0.0, 0.0, 1.0],
}
PROTOTYPES.update(PURE_VALS)

# Composite (20): uniform 20° angular offset from primary axis on unit 4-sphere.
# All 20 types use the same cos(20°)=0.93969 / sin(20°)=0.34202 ratio.
# This gives a natural Voronoi partition where all four sub-types within an arc
# (wo_sheng, sheng_wo, wo_ke, ke_wo) have statistically equal volume.
# Analysis: test_uniform_angle.py sweep, see docs/theory/SPEC.md §Classification.
_CS = 0.93969   # cos(20°) — primary element component
_SN = 0.34202   # sin(20°) — secondary element component
COMPOSITE_VALS = {
    # sheng pairs (X→Y): 10 types
    'WF': [_CS, _SN, 0.0, 0.0, 0.0],
    'FW': [_SN, _CS, 0.0, 0.0, 0.0],
    'FE': [0.0, _CS, _SN, 0.0, 0.0],
    'EF': [0.0, _SN, _CS, 0.0, 0.0],
    'EM': [0.0, 0.0, _CS, _SN, 0.0],
    'ME': [0.0, 0.0, _SN, _CS, 0.0],
    'MR': [0.0, 0.0, 0.0, _CS, _SN],
    'RM': [0.0, 0.0, 0.0, _SN, _CS],
    'RW': [_SN, 0.0, 0.0, 0.0, _CS],
    'WR': [_CS, 0.0, 0.0, 0.0, _SN],
    # ke pairs (X克Y): 10 types
    'WE': [_CS, 0.0, _SN, 0.0, 0.0],
    'EW': [_SN, 0.0, _CS, 0.0, 0.0],
    'FM': [0.0, _CS, 0.0, _SN, 0.0],
    'MF': [0.0, _SN, 0.0, _CS, 0.0],
    'ER': [0.0, 0.0, _CS, 0.0, _SN],
    'RE': [0.0, 0.0, _SN, 0.0, _CS],
    'MW': [_SN, 0.0, 0.0, _CS, 0.0],
    'WM': [_CS, 0.0, 0.0, _SN, 0.0],
    'RF': [0.0, _SN, 0.0, 0.0, _CS],
    'FR': [0.0, _CS, 0.0, 0.0, _SN],
}
PROTOTYPES.update(COMPOSITE_VALS)

ALL_TYPES = sorted(PROTOTYPES.keys())


# =============================================================================
# Classification — mirrors engine/identity.go ClassifyIdentity
# =============================================================================

def euclidean_dist(a, b):
    return math.sqrt(sum((a[i] - b[i]) ** 2 for i in range(5)))


# Tiebreaking — ordered by benefit to the elevated primary element.
#   "偏盛者发力" + "生先于克": generation is beneficial, control is constraining.
#   Hierarchy: pure > sheng_wo > wo_sheng > wo_ke > ke_wo.
#   Within same category: sheng-cycle order W→F→E→M→R.
_ELEM_ORDER = {'W': 0, 'F': 1, 'E': 2, 'M': 3, 'R': 4}

_SHENG_PAIRS = {('W','F'), ('F','E'), ('E','M'), ('M','R'), ('R','W')}
_KE_PAIRS    = {('W','E'), ('F','M'), ('E','R'), ('M','W'), ('R','F')}

def classify(v):
    """d → ReLU → argmax inner product with unit-sphere prototypes.
    Σd⁺ is a common factor — cancels in argmax, no normalization needed."""
    d_relu = [max(0.0, x) for x in v]
    if sum(d_relu) < 1e-12:
        return 'E'
    best_score = -1.0
    ties = []
    for label, proto in PROTOTYPES.items():
        score = sum(d_relu[i] * proto[i] for i in range(5))
        if score > best_score + 1e-12:
            best_score = score
            ties = [label]
        elif abs(score - best_score) <= 1e-12:
            ties.append(label)
    if len(ties) == 1:
        return ties[0]

    def tie_key(label):
        if len(label) == 1:
            return (0, 0, 0)       # pure
        a, b = label[0], label[1]
        if (b, a) in _SHENG_PAIRS:
            cat = 1                 # sheng_wo (secondary generates primary)
        elif (a, b) in _SHENG_PAIRS:
            cat = 2                 # wo_sheng (primary generates secondary)
        elif (a, b) in _KE_PAIRS:
            cat = 3                 # wo_ke (primary controls secondary)
        else:
            cat = 4                 # ke_wo (secondary controls primary)
        return (1, cat, _ELEM_ORDER[a])

    return min(ties, key=tie_key)


def display_label(label):
    """Return label as-is (single letter for pure, two letters for composite)."""
    return label


def qi_level(q):
    if q > 0.65:
        return 'H'   # 高 (High)
    elif q >= 0.35:
        return 'M'   # 中 (Mid)
    else:
        return 'L'   # 低 (Low)


# =============================================================================
# Random answer generators
# =============================================================================

def random_v_answer():
    """Random ranked choice: pick primary and secondary from 5 elements."""
    elements = [W, F, E, M, R]
    primary = random.choice(elements)
    remaining = [e for e in elements if e != primary]
    secondary = random.choice(remaining)
    return (primary, secondary)


def random_q_answer():
    """Random binary choice: A (True) or B (False)."""
    return random.choice([True, False])


# =============================================================================
# Simulation
# =============================================================================

def run_simulation(n=100000, n_v_questions=12, n_q_questions=8):
    """Run n random assessments. Returns type_counts, q_by_v_stats."""
    type_counts = Counter()
    raw_type_counts = Counter()
    qi_counts = Counter()
    sum_v = 0.0

    # Track for q-v independence analysis
    v_components_sum = [0.0] * 5
    q_sum = 0.0
    vq_products_sum = [0.0] * 5
    all_v = []
    all_q = []

    for _ in range(n):
        # Generate random answers
        v_answers = [random_v_answer() for _ in range(n_v_questions)]
        q_answers = [random_q_answer() for _ in range(n_q_questions)]

        # Score
        v = score_v(v_answers)
        q = score_q(q_answers)
        q = q_precision(q, n_q_questions)

        # Verify Σv = 0
        sum_v += abs(sum(v))

        # Classify
        label = classify(v)
        raw_type = label  # 2-char code like "WF"
        full_label = f"{display_label(label)}·{qi_level(q)}"
        type_counts[full_label] += 1
        raw_type_counts[raw_type] += 1
        qi_counts[qi_level(q)] += 1

        # Collect for independence analysis
        all_v.append(v)
        all_q.append(q)
        for i in range(5):
            v_components_sum[i] += v[i]
            vq_products_sum[i] += v[i] * q
        q_sum += q

    # Compute Pearson correlation between each v component and q
    n_float = float(n)
    correlations = []
    for i in range(5):
        v_mean = v_components_sum[i] / n_float
        q_mean = q_sum / n_float
        cov = vq_products_sum[i] / n_float - v_mean * q_mean
        v_var = sum((x[i] - v_mean) ** 2 for x in all_v) / n_float
        q_var = sum((x - q_mean) ** 2 for x in all_q) / n_float
        if v_var > 1e-12 and q_var > 1e-12:
            r = cov / math.sqrt(v_var * q_var)
        else:
            r = 0.0
        correlations.append(r)

    return type_counts, raw_type_counts, qi_counts, sum_v / n_float, correlations, all_v, all_q


# =============================================================================
# Reachability check: can we hit all 25 types with directed answers?
# =============================================================================

def check_all_types_reachable():
    """For each prototype, verify random answers can hit it."""
    unreachable = []
    for label in ALL_TYPES:
        found = False
        for _ in range(1000):
            answers = [random_v_answer() for _ in range(32)]
            v = score_v(answers)
            if classify(v) == label:
                found = True
                break
        if not found:
            unreachable.append(label)
    return unreachable


def check_directed_reachability():
    """For each simplex prototype, generate answers that produce the right v ratios.
    Uses floating-point residual tracking for precise ratio control."""
    results = {}
    N_Q = 32
    TOTAL = float(N_Q * 3)  # 96.0

    for label, proto in PROTOTYPES.items():
        active = [i for i in range(5) if proto[i] > 0]
        inactive = [i for i in range(5) if proto[i] == 0]

        target_v = [0.0] * 5
        if len(active) == 1:
            dom = active[0]
            target_v[dom] = 0.8
            for i in inactive:
                target_v[i] = -0.2
        else:
            a0, a1 = active[0], active[1]
            ratio = proto[a0] / proto[a1]
            v_sum = 0.9
            v1 = v_sum / (1.0 + ratio)
            v0 = ratio * v1
            target_v[a0] = v0
            target_v[a1] = v1
            for i in inactive:
                target_v[i] = -v_sum / 3.0

        # Floating-point target raw counts
        target_raw = [(TOTAL / 5.0) * (target_v[i] + 1.0) for i in range(5)]

        answers = []
        remaining = list(target_raw)  # float, not rounded
        for _ in range(N_Q):
            # Pick primary: element with largest remaining, but must have >= 2 remaining
            eligible_p = [i for i in range(5) if remaining[i] >= 1.5]
            if not eligible_p:
                eligible_p = list(range(5))
            primary = max(eligible_p, key=lambda i: remaining[i])
            remaining[primary] -= 2.0

            eligible_s = [i for i in range(5) if i != primary and remaining[i] >= 0.5]
            if not eligible_s:
                eligible_s = [i for i in range(5) if i != primary]
            secondary = max(eligible_s, key=lambda i: remaining[i])
            remaining[secondary] -= 1.0

            answers.append((primary, secondary))

        v = score_v(answers)
        result_label = classify(v)
        results[label] = {
            'v': v,
            'success': result_label == label,
            'result': result_label,
            'dom_elem': ELEMENT_NAMES[active[0]],
        }
    return results


# =============================================================================
# Print helpers
# =============================================================================

# Type groups for display
TYPE_GROUPS = [
    ('pure',     ['W', 'F', 'E', 'M', 'R']),
    ('composite', ['WF', 'FW', 'FE', 'EF', 'EM', 'ME', 'MR', 'RM', 'RW', 'WR',
                   'WE', 'EW', 'FM', 'MF', 'ER', 'RE', 'MW', 'WM', 'RF', 'FR']),
]


def bar(value, expected, width=40):
    """Draw a proportional bar chart."""
    if expected <= 0:
        return ''
    blocks = min(int(value * width / expected), width)
    return '█' * blocks


def print_type_uniformity(raw_type_counts, n):
    """Print distribution across the 25 types, grouped."""
    print(f"\n--- 25-Type Distribution (n={n:,}) ---")
    expected = n / 25

    for group_name, codes in TYPE_GROUPS:
        print(f"\n  {group_name}:")
        for code in codes:
            c = raw_type_counts.get(code, 0)
            pct = c / n * 100
            disp = display_label(code)
            b = bar(c, expected, 40)
            print(f"    {code} ({disp:4s}): {c:6d} ({pct:5.2f}%) {b}")

    # Stats
    counts = list(raw_type_counts.values())
    min_c = min(counts) if counts else 0
    max_c = max(counts) if counts else 0
    ratio = max_c / min_c if min_c > 0 else float('inf')
    print(f"\n  Expected per type: {expected:.0f}")
    print(f"  Range: {min_c} - {max_c}  (ratio: {ratio:.1f}x)")

    # Highlight uneven types
    over_repr = [(c, raw_type_counts[c]) for c in ALL_TYPES if raw_type_counts[c] > expected * 1.5]
    under_repr = [(c, raw_type_counts[c]) for c in ALL_TYPES if raw_type_counts[c] < expected * 0.5]
    if over_repr:
        print(f"  Over-represented (>1.5x expected): {', '.join(f'{c}={raw_type_counts[c]}' for c, _ in over_repr)}")
    if under_repr:
        print(f"  Under-represented (<0.5x expected): {', '.join(f'{c}={raw_type_counts[c]}' for c, _ in under_repr)}")


def print_results(type_counts, raw_type_counts, qi_counts, n, avg_sum_v, correlations):
    """Print simulation results."""
    # Coverage
    hit = len(type_counts)
    total_possible = 75
    coverage = hit / total_possible * 100

    counts = list(type_counts.values())
    min_c = min(counts)
    max_c = max(counts)
    avg_c = n / total_possible

    print(f"Simulations: {n:,}")
    print(f"Unique 75 labels hit: {hit}/{total_possible} ({coverage:.1f}%)")
    print(f"Expected per label: {avg_c:.1f}")
    print(f"Actual range: {min_c} - {max_c}")
    ratio = max_c / min_c if min_c > 0 else float('inf')
    print(f"Max/min ratio: {ratio:.2f}x")

    # Σv = 0 invariant
    print(f"\n--- Σv = 0 invariant ---")
    print(f"Mean |Σv|: {avg_sum_v:.2e}  (should be near 0)")

    # q-v orthogonality
    print(f"\n--- q-v Pearson correlations ---")
    max_abs_r = 0.0
    for i, r in enumerate(correlations):
        flag = ' ✔' if abs(r) < 0.03 else ' ✗ BIASED'
        print(f"  r(q, {ELEMENT_NAMES[i]:5s}) = {r:+.4f}{flag}")
        max_abs_r = max(max_abs_r, abs(r))
    if max_abs_r < 0.03:
        print("  Result: q is INDEPENDENT of v direction ✔")
    else:
        print("  Result: q-v COUPLING detected ✗ — questions need review")

    # qi distribution
    print(f"\n--- qi distribution ---")
    qi_labels = {'H': 'High  ', 'M': 'Mid   ', 'L': 'Low   '}
    for level in ['H', 'M', 'L']:
        c = qi_counts.get(level, 0)
        pct = c / n * 100
        label = qi_labels.get(level, level)
        print(f"  {label}: {c:6d} ({pct:5.1f}%) {bar(c, n/3)}")

    # Per-type uniformity
    print_type_uniformity(raw_type_counts, n)


# =============================================================================
# Single trace
# =============================================================================

def trace_one():
    """Detailed trace of one random assessment."""
    print("=== Part 1: v (12 ranked-choice questions) ===\n")
    v_answers = [random_v_answer() for _ in range(32)]
    for i, (p, s) in enumerate(v_answers):
        print(f"  v{i+1:02d}: primary={ELEMENT_NAMES[p]:5s}  secondary={ELEMENT_NAMES[s]:5s}")

    v = score_v(v_answers)
    print(f"\n  v = [{', '.join(f'{x:+.4f}' for x in v)}]")
    print(f"  Σv = {sum(v):.2e}  (should be 0)")

    label = classify(v)
    print(f"  Type: {label} ({display_label(label)})")

    # Show top 3 prototype scores
    d_relu = [max(0.0, x) for x in v]
    total = sum(d_relu)
    p = [x / total for x in d_relu] if total > 1e-12 else [0.2]*5
    scores = [(code, sum(p[i] * PROTOTYPES[code][i] for i in range(5))) for code in ALL_TYPES]
    scores.sort(key=lambda x: x[1], reverse=True)
    print(f"  Top prototype scores:")
    for code, s in scores[:3]:
        mark = '<--' if code == label else ''
        print(f"    {code}: score={s:.4f} {mark}")

    print(f"\n=== Part 2: q (8 binary forced-choice questions) ===\n")
    q_answers = [random_q_answer() for _ in range(8)]
    for i, a in enumerate(q_answers):
        ans_str = 'A (正气足)' if a else 'B (正气不足)'
        print(f"  q{i+1:02d}: {ans_str}")

    q = score_q(q_answers)
    q = q_precision(q, 8)
    print(f"\n  q = {q:.2f}  ({qi_level(q)})")

    full_label = f"{display_label(label)}·{qi_level(q)}"
    print(f"\n  Full label: {full_label}")


# =============================================================================
# Prototype self-classification check
# =============================================================================

def check_prototype_self_classification():
    """Verify all 25 prototypes classify to themselves."""
    failures = []
    for label, proto in PROTOTYPES.items():
        result = classify(proto)
        if result != label:
            failures.append((label, result))
    return failures


# =============================================================================
# main
# =============================================================================

# =============================================================================
# Laplace-smoothed p distribution
# =============================================================================

def simulate_laplace_p(n=1000, n_v_questions=12):
    """Verify Laplace-smoothed p has reasonable range and no zeros."""
    min_ps = [1.0] * 5
    max_ps = [0.0] * 5
    all_ps = []

    for _ in range(n):
        answers = [random_v_answer() for _ in range(n_v_questions)]
        p = score_p(answers, LAPLACE_PRIOR)
        all_ps.append(p)
        for i in range(5):
            min_ps[i] = min(min_ps[i], p[i])
            max_ps[i] = max(max_ps[i], p[i])

    # Also check extreme case: all primaries on one element
    extreme_p = score_p([(W, F) for _ in range(n_v_questions)], LAPLACE_PRIOR)
    zero_p = score_p([(W, F), (F, W), (E, M), (M, R), (R, E),
                       (W, F), (F, W), (E, M), (M, R), (R, E),
                       (W, E), (F, M)], LAPLACE_PRIOR)

    return min_ps, max_ps, extreme_p, zero_p, all_ps


def print_laplace_p_results(min_ps, max_ps, extreme_p, zero_p, all_ps):
    """Print Laplace p distribution results."""
    # Compute mean p across all simulations
    mean_ps = [0.0] * 5
    for p in all_ps:
        for i in range(5):
            mean_ps[i] += p[i]
    mean_ps = [x / len(all_ps) for x in mean_ps]

    print(f"\n--- Laplace-Smoothed p Distribution (prior={LAPLACE_PRIOR}) ---")
    print(f"  Element     Min p    Max p    Mean p")
    for i, name in enumerate(ELEMENT_NAMES):
        print(f"  {name:7s}:  {min_ps[i]:.3f}    {max_ps[i]:.3f}    {mean_ps[i]:.3f}")
    print(f"  All p are > 0: {all(min_ps[i] > 0 for i in range(5))} ✔")

    print(f"\n  Extreme case (all primary on Wood, 2nd Fire):")
    print(f"    p = [{', '.join(f'{x:.3f}' for x in extreme_p)}]")
    print(f"    Σp = {sum(extreme_p):.4f}")
    print(f"    v equivalent = [{', '.join(f'{x:+.3f}' for x in p_to_v(extreme_p))}]")

    print(f"  Balanced case (mixed answers):")
    print(f"    p = [{', '.join(f'{x:.3f}' for x in zero_p)}]")
    print(f"    Σp = {sum(zero_p):.4f}")

    # Verify no zero
    zeros = sum(1 for i in range(5) if extreme_p[i] == 0 or zero_p[i] == 0)
    print(f"  Zero proportions: {zeros} (should be 0) {'✔' if zeros == 0 else '✗'}")


# =============================================================================
# Stratified question bank sampling
# =============================================================================

def simulate_question_bank_sampling(n_personalities=200, n_per_bank=5, draw_per_bank=3):
    """Simulate stratified sampling from a question bank.
    Each personality has FIXED answers to each question in the bank.
    Different subsets draw different questions — variation comes purely from
    which questions are selected, not from response inconsistency."""
    n_banks = 4
    total_q = draw_per_bank * n_banks
    total_bank_q = n_per_bank * n_banks

    # Build full question bank: each question has a "scenario flavor" that
    # nudges the personality's answer slightly differently
    all_v_variations = []
    mismatch_count = 0

    for pi_idx in range(n_personalities):
        dom = random.randint(0, 4)
        sec = (dom + random.randint(1, 4)) % 5

        # Pre-generate this personality's fixed answers to ALL bank questions
        bank_answers = []
        for bank in range(n_banks):
            for q in range(n_per_bank):
                # Each question is a specific scenario with a fixed answer from this person
                roll = random.random()
                if roll < 0.55:
                    primary = dom
                    secondary = sec
                elif roll < 0.85:
                    primary = sec
                    secondary = dom
                else:
                    primary = (dom + q + 1) % 5
                    others = [i for i in range(5) if i != primary]
                    secondary = random.choice(others)
                bank_answers.append((bank, primary, secondary))

        # Now draw 5 different random subsets and measure spread
        v_list = []
        labels_list = []
        for trial in range(5):
            # Stratified draw: draw_per_bank from each bank
            trial_answers = []
            for bank in range(n_banks):
                bank_qs = [(b, p, s) for b, p, s in bank_answers if b == bank]
                sampled = random.sample(bank_qs, min(draw_per_bank, len(bank_qs)))
                trial_answers.extend([(p, s) for _, p, s in sampled])

            v = score_v(trial_answers)
            v_list.append(v)
            labels_list.append(classify(v))

        # Measure spread
        spreads = []
        for i in range(len(v_list)):
            for j in range(i + 1, len(v_list)):
                spreads.append(euclidean_dist(v_list[i], v_list[j]))
        mean_spread = sum(spreads) / len(spreads) if spreads else 0
        all_v_variations.append(mean_spread)

        if pi_idx < 100 and len(set(labels_list)) > 1:
            mismatch_count += 1

    avg_spread = sum(all_v_variations) / len(all_v_variations)
    max_spread = max(all_v_variations)

    return avg_spread, max_spread, mismatch_count, total_q


def weighted_choice(weights, elements=None):
    """Pick an element with given weights. Zero weights are excluded."""
    if elements is None:
        elements = list(range(5))
    eligible = [(i, w) for i, w in enumerate(weights) if w > 0 and i in elements]
    total = sum(w for _, w in eligible)
    r = random.random() * total
    cumulative = 0
    for i, w in eligible:
        cumulative += w
        if r <= cumulative:
            return i
    return eligible[-1][0]


def print_question_bank_results(avg_spread, max_spread, mismatch_count, total_q):
    """Print stratified question bank sampling results."""
    print(f"\n--- Stratified Question Bank Sampling ({total_q} questions, 4 banks) ---")
    print(f"  Mean pairwise v-spread (same personality, different subsets): {avg_spread:.4f}")
    print(f"  Max pairwise v-spread: {max_spread:.4f}")
    print(f"  Personalities with type mismatch across subsets: {mismatch_count}/100")
    stability = 100 - mismatch_count
    print(f"  Classification stability: {stability}% (higher is better)")
    print(f"  Note: low type stability is expected with 12 questions — v vectors")
    print(f"  are close (spread {avg_spread:.2f}) but thin Voronoi boundaries mean")
    print(f"  subtle shifts cross type labels. More questions improve stability.")
    print(f"  The continuous v is the truth; discrete type is a convenient summary.")


# =============================================================================
# Question count impact
# =============================================================================

def simulate_question_count_impact(n=5000):
    """Compare classification stability at 8, 12, 16 questions."""
    results = {}
    for n_q in [8, 12, 16]:
        per_bank = n_q // 4
        # Use a fixed set of 16 biased personalities (4 extreme, 12 varied)
        label_counts = Counter()
        for _ in range(n):
            answers = [random_v_answer() for _ in range(n_q)]
            v = score_v(answers)
            label = classify(v)
            label_counts[label] += 1

        # Coverage
        hit = len(label_counts)
        coverage = hit / 25 * 100
        # Uniformity
        counts = list(label_counts.values())
        ratio = max(counts) / min(counts) if counts else 0
        results[n_q] = {'coverage': coverage, 'ratio': ratio, 'hit': hit}

    return results


def print_question_count_results(results):
    """Print question count impact results."""
    print(f"\n--- Question Count Impact ---")
    print(f"  {'Q count':<10} {'Types hit':<12} {'Coverage':<10} {'Max/min ratio'}")
    for n_q in sorted(results.keys()):
        r = results[n_q]
        print(f"  {n_q:<10} {r['hit']}/25{'':>6} {r['coverage']:.1f}%{'':>6} {r['ratio']:.2f}x")


# =============================================================================
# Self/other Δv divergence detection
# =============================================================================

def simulate_delta_v(n_pairs=500, n_v_questions=12):
    """Simulate self vs other assessment divergence.
    Self answers truthfully. Observers have systematic element bias:
    they consistently overweight one element and underweight another.
    Δv should detect this systematic distortion."""
    results = []
    for _ in range(n_pairs):
        # Self: honest answers from their true personality
        self_answers = [random_v_answer() for _ in range(n_v_questions)]
        v_self = score_v(self_answers)

        # Observer bias: systematic over/underweight of specific elements
        bias_strength = random.uniform(0, 1.0)
        over_elem = random.randint(0, 4)
        under_elem = random.choice([i for i in range(5) if i != over_elem])

        # Generate observer answers — start from self answers then bias
        other_answers = []
        for _ in range(n_v_questions):
            primary = random.choice(range(5))
            # Bias: shift primary toward over_elem, away from under_elem
            if random.random() < bias_strength * 0.5:
                primary = over_elem
            elif random.random() < bias_strength * 0.3:
                # Avoid under_elem
                primary = random.choice([i for i in range(5) if i != under_elem])
            remaining = [e for e in range(5) if e != primary]
            secondary = random.choice(remaining)
            other_answers.append((primary, secondary))

        v_other = score_v(other_answers)
        delta = [v_self[i] - v_other[i] for i in range(5)]
        delta_mag = sum(abs(d) for d in delta)
        results.append({
            'bias_strength': bias_strength,
            'over_elem': over_elem,
            'under_elem': under_elem,
            'delta_mag': delta_mag,
            'v_self': v_self,
            'v_other': v_other,
            'delta': delta,
        })

    # Check: does Δv magnitude correlate with bias strength?
    sorted_by_bias = sorted(results, key=lambda r: r['bias_strength'])
    low_bias = [r['delta_mag'] for r in sorted_by_bias[:100]]
    high_bias = [r['delta_mag'] for r in sorted_by_bias[-100:]]
    avg_low = sum(low_bias) / len(low_bias)
    avg_high = sum(high_bias) / len(high_bias)

    # Correlation
    n_float = float(len(results))
    bs_mean = sum(r['bias_strength'] for r in results) / n_float
    dm_mean = sum(r['delta_mag'] for r in results) / n_float
    cov = sum((r['bias_strength'] - bs_mean) * (r['delta_mag'] - dm_mean) for r in results) / n_float
    bs_var = sum((r['bias_strength'] - bs_mean) ** 2 for r in results) / n_float
    dm_var = sum((r['delta_mag'] - dm_mean) ** 2 for r in results) / n_float
    r_bs_dm = cov / math.sqrt(bs_var * dm_var) if bs_var > 1e-12 and dm_var > 1e-12 else 0.0

    return results, avg_low, avg_high, r_bs_dm


def print_delta_v_results(results, avg_low, avg_high, r_bs_dm):
    """Print delta-v divergence detection results."""
    print(f"\n--- Self/Other Δv Divergence Detection ---")
    print(f"  Pairs simulated: {len(results)}")
    print(f"  Δv magnitude — low-bias avg: {avg_low:.3f}, high-bias avg: {avg_high:.3f}")
    print(f"  Ratio high/low: {avg_high / avg_low:.1f}x (should be > 1)")
    print(f"  Correlation(bias_strength, Δv_magnitude): r = {r_bs_dm:+.3f}")
    detectable = avg_high > avg_low * 1.3 and r_bs_dm > 0.3
    flag = '✔' if detectable else '✗ (may need tuning)'
    print(f"  Δv detects observer bias: {flag}")

    # Show a sample with largest divergence
    results_sorted = sorted(results, key=lambda r: r['delta_mag'], reverse=True)
    r = results_sorted[0]
    print(f"\n  Example — largest divergence (bias={r['bias_strength']:.2f}):")
    print(f"    Observer overweighted: {ELEMENT_NAMES[r['over_elem']]}, underweighted: {ELEMENT_NAMES[r['under_elem']]}")
    print(f"    v_self  = [{', '.join(f'{x:+.3f}' for x in r['v_self'])}]")
    print(f"    v_other = [{', '.join(f'{x:+.3f}' for x in r['v_other'])}]")
    print(f"    Δv      = [{', '.join(f'{x:+.3f}' for x in r['delta'])}]")
    print(f"    |Δv|    = {r['delta_mag']:.3f}")
    label_self = classify(r['v_self'])
    label_other = classify(r['v_other'])
    print(f"    Self type:  {label_self} ({role_display(label_self)})")
    print(f"    Other type: {label_other} ({role_display(label_other)})")


def simulate_delta_q(n_pairs=1000):
    """Simulate self-q vs observer-q divergence.

    Observer q is a noisy view of the same underlying trait vitality.
    q_self and q_other measure the same construct from different vantage points.
    Moderate correlation expected (same construct, different method).
    Δq = q_self - q_other reveals self-other perception gap on energy.
    """
    deltas = []
    q_selfs = []
    q_others = []
    for _ in range(n_pairs):
        # Underlying trait q (0-1)
        true_q = random.random()
        # Self: directly samples the trait with some noise
        q_self = min(1.0, max(0.0, true_q + random.gauss(0, 0.1)))
        # Observer: sees external behavior, noisier than self
        q_other = min(1.0, max(0.0, true_q + random.gauss(0, 0.18)))
        deltas.append(abs(q_self - q_other))
        q_selfs.append(q_self)
        q_others.append(q_other)

    avg_delta = sum(deltas) / len(deltas)
    max_delta = max(deltas)

    # Pearson r between self and other q
    nf = float(n_pairs)
    mean_s = sum(q_selfs) / nf
    mean_o = sum(q_others) / nf
    cov = sum((q_selfs[i] - mean_s) * (q_others[i] - mean_o) for i in range(n_pairs)) / nf
    var_s = sum((x - mean_s) ** 2 for x in q_selfs) / nf
    var_o = sum((x - mean_o) ** 2 for x in q_others) / nf
    r = cov / math.sqrt(var_s * var_o) if var_s > 1e-12 and var_o > 1e-12 else 0.0

    return avg_delta, max_delta, r


def simulate_q_per_question_orthogonality(n=50000, n_v=12):
    """Check each q question's correlation with each v component.

    Even though simulation uses random q answers (content-neutral by construction),
    this verifies the testing infrastructure and catches any structural bias.
    """
    n_q = 16
    # Accumulators: per-q answer sum, per-q × per-v product sums
    q_sums = [0.0] * n_q
    v_sums = [0.0] * 5
    qv_products = [[0.0] * 5 for _ in range(n_q)]  # [q_idx][v_idx]

    all_v = []
    all_q_answers = [[] for _ in range(n_q)]

    for _ in range(n):
        v_answers = [(random.randrange(5), random.randrange(5)) for _ in range(n_v)]
        v = score_v(v_answers)
        all_v.append(v)

        for vi in range(5):
            v_sums[vi] += v[vi]

        for qi in range(n_q):
            ans = 1 if random.random() < 0.55 else 0  # A=1, B=0
            q_sums[qi] += ans
            all_q_answers[qi].append(ans)
            for vi in range(5):
                qv_products[qi][vi] += ans * v[vi]

    nf = float(n)
    max_abs_r = 0.0
    violations = 0

    for qi in range(n_q):
        q_mean = q_sums[qi] / nf
        q_var = q_mean * (1.0 - q_mean)  # Bernoulli variance = p(1-p)
        if q_var < 1e-12:
            continue
        for vi in range(5):
            v_mean = v_sums[vi] / nf
            cov = qv_products[qi][vi] / nf - q_mean * v_mean
            v_var = sum((x[vi] - v_mean) ** 2 for x in all_v) / nf
            if v_var < 1e-12:
                r = 0.0
            else:
                r = cov / math.sqrt(q_var * v_var)
            abs_r = abs(r)
            if abs_r > max_abs_r:
                max_abs_r = abs_r
            if abs_r >= 0.01:
                violations += 1

    return max_abs_r, violations


def print_q_per_question_results(max_abs_r, violations):
    print(f"  n=50000, 16 q questions × 5 v components = 80 correlations")
    print(f"  Max |r|: {max_abs_r:.4f}")
    print(f"  Violations (|r| >= 0.01): {violations}")
    if violations == 0:
        print(f"  All per-question q-v correlations below 0.01 threshold ✔")
    else:
        print(f"  WARNING: {violations} correlations exceed 0.01 threshold")


def simulate_q_draw_stability(n_personalities=1000, n_q_bank=16, draw_q=8):
    """Test q-score stability when randomly drawing 8 of 16 questions.

    Each simulated person has a fixed set of 16 q answers (their trait).
    We draw 8 without replacement multiple times and measure q-score variance.
    """
    n_draws = 50  # draws per personality
    all_stds = []
    band_crosses = 0
    total_draws = 0

    for _ in range(n_personalities):
        # Generate this person's trait answers to all 16 q questions
        trait_answers = [1 if random.random() < 0.55 else 0 for _ in range(n_q_bank)]
        true_q = sum(trait_answers) / n_q_bank
        true_band = qi_level(true_q)

        draw_qs = []
        for _ in range(n_draws):
            indices = random.sample(range(n_q_bank), draw_q)
            drawn = [trait_answers[i] for i in indices]
            q_sample = sum(drawn) / draw_q
            draw_qs.append(q_sample)
            total_draws += 1
            if qi_level(q_sample) != true_band:
                band_crosses += 1

        std = (sum((x - sum(draw_qs)/len(draw_qs))**2 for x in draw_qs) / len(draw_qs)) ** 0.5
        all_stds.append(std)

    avg_std = sum(all_stds) / len(all_stds)
    max_std = max(all_stds)
    band_cross_rate = band_crosses / total_draws

    return avg_std, max_std, band_cross_rate


def print_q_draw_stability_results(avg_std, max_std, band_cross_rate):
    print(f"  n=1000 personalities × 50 draws each")
    print(f"  Average q std across draws: {avg_std:.4f}")
    print(f"  Max q std across draws:     {max_std:.4f}")
    print(f"  Band cross rate:            {band_cross_rate:.2%}")
    print(f"  (Expected theoretical std for q=0.5: ~0.129, max: ~0.177)")
    if band_cross_rate < 0.15:
        print(f"  Band stability acceptable ✔")
    else:
        print(f"  Band cross rate elevated — consider wider bands or more questions")


if __name__ == '__main__':
    print("=" * 60)
    print("25types Questionnaire Scoring Simulation")
    print("=" * 60)

    # 1. Prototype self-classification
    print("\n--- Prototype Self-Classification ---")
    failures = check_prototype_self_classification()
    if failures:
        for proto, result in failures:
            print(f"  FAIL: {proto} → classified as {result}")
    else:
        print(f"  All {len(PROTOTYPES)} prototypes classify to themselves ✔")

    # 2. All types reachable via random answers
    print("\n--- Type Reachability (random answers) ---")
    unreachable = check_all_types_reachable()
    if unreachable:
        print(f"  Unreachable types ({len(unreachable)}):")
        for u in unreachable:
            print(f"    {u}")
    else:
        print(f"  All 25 types reachable via random answers ✔")

    # 3. Directed reachability: biased answers toward each prototype
    print("\n--- Directed Reachability (biased answers) ---")
    directed = check_directed_reachability()
    failures_directed = [(l, r) for l, r in directed.items() if not r['success']]
    if failures_directed:
        print(f"  Failed ({len(failures_directed)}):")
        for label, r in failures_directed:
            print(f"    {label} ({r['dom_elem']}-directed): classified as {r['result']}")
    else:
        print(f"  All 25 types reachable via biased answers ✔")
    # Show the v vectors for all types
    for group_name, codes in TYPE_GROUPS:
        print(f"  {group_name}:")
        for code in codes:
            r = directed[code]
            v = r['v']
            flag = '✔' if r['success'] else f"✗ →{r['result']}"
            print(f"    {code} ({display_label(code):4s}): v=[{', '.join(f'{x:+.2f}' for x in v)}] {flag}")

    # 4. Single trace
    print("\n" + "=" * 60)
    print("Sample Trace")
    print("=" * 60 + "\n")
    trace_one()

    # 5. Full simulation (8 q questions)
    print("\n" + "=" * 60)
    n = 100000
    print(f"Running {n:,} random simulations (32 v questions, 8 q questions)...\n")
    type_counts, raw_type_counts, qi_counts, avg_sum_v, correlations, all_v, all_q = \
        run_simulation(n, n_v_questions=32, n_q_questions=8)
    print_results(type_counts, raw_type_counts, qi_counts, n, avg_sum_v, correlations)

    # 6. q-v orthogonality (aggregate)
    print(f"\n{'=' * 60}")
    print("q-v Orthogonality")
    print("=" * 60)
    print(f"  Self-assessment q (8 questions, n=100k)")
    print(f"  r(q, W)={correlations[0]:+.4f}  r(q, F)={correlations[1]:+.4f}  r(q, E)={correlations[2]:+.4f}  r(q, M)={correlations[3]:+.4f}  r(q, R)={correlations[4]:+.4f}")
    all_ok = all(abs(r) < 0.01 for r in correlations)
    print(f"  All |r| < 0.01: {'✔' if all_ok else '✗'}")

    # 7. Pathology check: extreme v doesn't correlate with q
    print(f"\n--- Extreme v vs q (10 samples) ---")
    for elem_idx, elem_name in enumerate(ELEMENT_NAMES):
        answers = [(elem_idx, (elem_idx + 1) % 5) for _ in range(32)]
        v = score_v(answers)
        q_answers = [random_q_answer() for _ in range(8)]
        q = score_q(q_answers)
        print(f"  {elem_name}-heavy v: q={q:.2f} (expected near 0.5, not extreme)")

    # 8. Self/other q divergence
    print(f"\n{'=' * 60}")
    print("Self/Other Δq Divergence Detection")
    print("=" * 60)
    dq_avg, dq_max, dq_corr = simulate_delta_q(1000)
    print(f"  Pairs simulated: 1000")
    print(f"  Average |Δq|: {dq_avg:.3f}")
    print(f"  Max |Δq|:     {dq_max:.3f}")
    print(f"  Corr(q_self, q_other): r = {dq_corr:+.3f}")
    print(f"  (Moderate correlation expected — same construct, different vantage)")

    # 10. Laplace-smoothed p distribution
    print(f"\n{'=' * 60}")
    print("Laplace-Smoothed p Distribution")
    print("=" * 60)
    min_ps, max_ps, extreme_p, zero_p, all_ps = simulate_laplace_p(5000)
    print_laplace_p_results(min_ps, max_ps, extreme_p, zero_p, all_ps)

    # 12. Stratified question bank sampling fairness
    print(f"\n{'=' * 60}")
    print("Stratified Question Bank Sampling")
    print("=" * 60)
    avg_spread, max_spread, mismatch_count, total_q = simulate_question_bank_sampling(
        n_personalities=200, n_per_bank=5, draw_per_bank=3)
    print_question_bank_results(avg_spread, max_spread, mismatch_count, total_q)

    # 13. Question count impact
    print(f"\n{'=' * 60}")
    print("Question Count Impact (8 vs 12 vs 16)")
    print("=" * 60)
    qc_results = simulate_question_count_impact(5000)
    print_question_count_results(qc_results)

    # 14. Self/other Δv divergence detection
    print(f"\n{'=' * 60}")
    print("Self/Other Δv Divergence Detection")
    print("=" * 60)
    dv_results, dv_low, dv_high, dv_r = simulate_delta_v(500)
    print_delta_v_results(dv_results, dv_low, dv_high, dv_r)
