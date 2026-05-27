"""25types L2 Profile Model — Full Verification

Verifies the L2 unit-sphere profile model against the current simplex model:
  1. L2 normalization preserves element ordering
  2. d+ activation equivalence (L2 vs simplex)
  3. Classification consistency (25-type distribution, uniformity)
  4. Prototype self-classification
  5. 顺逆平 (synastry) distribution across 300 pairs
  6. Bond output consistency
  7. Flow scale recalibration
"""

import math
import random
from collections import Counter, defaultdict
import sys

# =============================================================================
# Constants
# =============================================================================

W, F, E, M, R = 0, 1, 2, 3, 4
ELEMENT_NAMES = ['Wood', 'Fire', 'Earth', 'Metal', 'Water']
ELEMENT_CODES = ['W', 'F', 'E', 'M', 'R']

# L2 neutral point: all elements equal, normalized to unit length
NEUTRAL_L2 = 1.0 / math.sqrt(5)  # ≈ 0.44721

# S (生) matrix
S = [
    [0, 1, 0, 0, 0],
    [0, 0, 1, 0, 0],
    [0, 0, 0, 1, 0],
    [0, 0, 0, 0, 1],
    [1, 0, 0, 0, 0],
]

# C (克) matrix
C = [
    [0, 0, 1, 0, 0],
    [0, 0, 0, 1, 0],
    [0, 0, 0, 0, 1],
    [1, 0, 0, 0, 0],
    [0, 1, 0, 0, 0],
]

# Prototypes on L2 unit sphere (cos42° ≈ 0.74314, sin42° ≈ 0.66913)
PROTOTYPES = {
    'W':  [1.00000, 0.00000, 0.00000, 0.00000, 0.00000],
    'F':  [0.00000, 1.00000, 0.00000, 0.00000, 0.00000],
    'E':  [0.00000, 0.00000, 1.00000, 0.00000, 0.00000],
    'M':  [0.00000, 0.00000, 0.00000, 1.00000, 0.00000],
    'R':  [0.00000, 0.00000, 0.00000, 0.00000, 1.00000],
    'WF': [0.74314, 0.66913, 0.00000, 0.00000, 0.00000],
    'FW': [0.66913, 0.74314, 0.00000, 0.00000, 0.00000],
    'FE': [0.00000, 0.74314, 0.66913, 0.00000, 0.00000],
    'EF': [0.00000, 0.66913, 0.74314, 0.00000, 0.00000],
    'EM': [0.00000, 0.00000, 0.74314, 0.66913, 0.00000],
    'ME': [0.00000, 0.00000, 0.66913, 0.74314, 0.00000],
    'MR': [0.00000, 0.00000, 0.00000, 0.74314, 0.66913],
    'RM': [0.00000, 0.00000, 0.00000, 0.66913, 0.74314],
    'RW': [0.66913, 0.00000, 0.00000, 0.00000, 0.74314],
    'WR': [0.74314, 0.00000, 0.00000, 0.00000, 0.66913],
    'WE': [0.74314, 0.00000, 0.66913, 0.00000, 0.00000],
    'EW': [0.66913, 0.00000, 0.74314, 0.00000, 0.00000],
    'FM': [0.00000, 0.74314, 0.00000, 0.66913, 0.00000],
    'MF': [0.00000, 0.66913, 0.00000, 0.74314, 0.00000],
    'ER': [0.00000, 0.00000, 0.74314, 0.00000, 0.66913],
    'RE': [0.00000, 0.00000, 0.66913, 0.00000, 0.74314],
    'MW': [0.66913, 0.00000, 0.00000, 0.74314, 0.00000],
    'WM': [0.74314, 0.00000, 0.00000, 0.66913, 0.00000],
    'RF': [0.00000, 0.66913, 0.00000, 0.00000, 0.74314],
    'FR': [0.00000, 0.74314, 0.00000, 0.00000, 0.66913],
}

ALL_TYPES = sorted(PROTOTYPES.keys())


# =============================================================================
# Scoring functions
# =============================================================================

def score_simplex_v(answers):
    """Current model: v (Σ=0) from forced-choice answers."""
    raw = [0, 0, 0, 0, 0]
    for p, s in answers:
        raw[p] += 2
        raw[s] += 1
    total = sum(raw)
    if total == 0:
        return [0.0, 0.0, 0.0, 0.0, 0.0]
    return [(raw[i] / total - 0.2) * 5.0 for i in range(5)]


def score_simplex_p(answers):
    """Current model: p (Σ=1) display proportions."""
    raw = [0, 0, 0, 0, 0]
    for p, s in answers:
        raw[p] += 2
        raw[s] += 1
    total = sum(raw)
    if total == 0:
        return [0.2, 0.2, 0.2, 0.2, 0.2]
    return [raw[i] / total for i in range(5)]


def score_l2_p(answers):
    """L2 model: p (Σx²=1) from forced-choice answers.
    First compute raw proportions, then L2-normalize."""
    raw = [0, 0, 0, 0, 0]
    for p, s in answers:
        raw[p] += 2
        raw[s] += 1
    total = sum(raw)
    if total == 0:
        return [NEUTRAL_L2] * 5
    # Raw proportions (simplex)
    f = [raw[i] / total for i in range(5)]
    # L2 normalize
    norm = math.sqrt(sum(x * x for x in f))
    if norm < 1e-12:
        return [NEUTRAL_L2] * 5
    return [f[i] / norm for i in range(5)]


def p_to_v_l2(p):
    """L2: compute deviation from neutral."""
    return [p[i] - NEUTRAL_L2 for i in range(5)]


def relu(vec):
    return [max(0.0, x) for x in vec]


def dot(a, b):
    return sum(a[i] * b[i] for i in range(5))


# =============================================================================
# Classification
# =============================================================================

_SHENG_PAIRS = {('W','F'), ('F','E'), ('E','M'), ('M','R'), ('R','W')}
_KE_PAIRS    = {('W','E'), ('F','M'), ('E','R'), ('M','W'), ('R','F')}
_ELEM_ORDER = {'W': 0, 'F': 1, 'E': 2, 'M': 3, 'R': 4}


def classify_simplex(v):
    """Current model: argmax ReLU(v) · prototype."""
    d_relu = relu(v)
    if sum(d_relu) < 1e-12:
        return 'E'
    best_score = -1.0
    ties = []
    for label, proto in PROTOTYPES.items():
        score = dot(d_relu, proto)
        if score > best_score + 1e-12:
            best_score = score
            ties = [label]
        elif abs(score - best_score) <= 1e-12:
            ties.append(label)
    if len(ties) == 1:
        return ties[0]

    def tie_key(label):
        if len(label) == 1:
            return (0, 0, 0)
        a, b = label[0], label[1]
        if (b, a) in _SHENG_PAIRS:
            cat = 1
        elif (a, b) in _SHENG_PAIRS:
            cat = 2
        elif (a, b) in _KE_PAIRS:
            cat = 3
        else:
            cat = 4
        return (1, cat, _ELEM_ORDER[a])
    return min(ties, key=tie_key)


def classify_l2(p):
    """L2 model: argmax ReLU(p - neutral) · prototype."""
    v = p_to_v_l2(p)
    d_relu = relu(v)
    if sum(d_relu) < 1e-12:
        return 'E'
    best_score = -1.0
    ties = []
    for label, proto in PROTOTYPES.items():
        score = dot(d_relu, proto)
        if score > best_score + 1e-12:
            best_score = score
            ties = [label]
        elif abs(score - best_score) <= 1e-12:
            ties.append(label)
    if len(ties) == 1:
        return ties[0]

    def tie_key(label):
        if len(label) == 1:
            return (0, 0, 0)
        a, b = label[0], label[1]
        if (b, a) in _SHENG_PAIRS:
            cat = 1
        elif (a, b) in _SHENG_PAIRS:
            cat = 2
        elif (a, b) in _KE_PAIRS:
            cat = 3
        else:
            cat = 4
        return (1, cat, _ELEM_ORDER[a])
    return min(ties, key=tie_key)


# =============================================================================
# 顺逆平 (Synastry) computation
# =============================================================================

def mat_vec_mul(M, v):
    return [sum(M[i][j] * v[j] for j in range(5)) for i in range(5)]


# Pre-compute symmetric S+S^T and C+C^T
SpSt = [[S[i][j] + S[j][i] for j in range(5)] for i in range(5)]
CpCt = [[C[i][j] + C[j][i] for j in range(5)] for i in range(5)]


def compute_synastry(pA, pB, eps=0.01):
    """Compute 顺(shun)/逆(ni)/平(ping) between two L2 profiles."""
    sh = dot(pA, mat_vec_mul(SpSt, pB))
    ke = dot(pA, mat_vec_mul(CpCt, pB))
    if sh > ke + eps:
        return '顺', sh, ke
    elif ke > sh + eps:
        return '逆', sh, ke
    else:
        return '平', sh, ke


# =============================================================================
# Random answer generators
# =============================================================================

def random_v_answer():
    elements = [W, F, E, M, R]
    primary = random.choice(elements)
    remaining = [e for e in elements if e != primary]
    secondary = random.choice(remaining)
    return (primary, secondary)


# =============================================================================
# Test 1: Element ordering preservation
# =============================================================================

def test_ordering_preservation(n=5000, n_q=32):
    """Verify L2 normalization preserves element ranking order."""
    swaps = 0
    total_pairs = 0
    for _ in range(n):
        answers = [random_v_answer() for _ in range(n_q)]
        p_simplex = score_simplex_p(answers)
        p_l2 = score_l2_p(answers)
        # Check pairwise ordering
        for i in range(5):
            for j in range(i + 1, 5):
                total_pairs += 1
                sim_order = 1 if p_simplex[i] > p_simplex[j] else (-1 if p_simplex[i] < p_simplex[j] else 0)
                l2_order = 1 if p_l2[i] > p_l2[j] else (-1 if p_l2[i] < p_l2[j] else 0)
                if sim_order != l2_order and sim_order != 0 and l2_order != 0:
                    swaps += 1
    return swaps, total_pairs


# =============================================================================
# Test 2: d+ activation equivalence
# =============================================================================

def test_dplus_equivalence(n=5000, n_q=32):
    """Verify L2 d+ (ReLU(p - neutral)) and simplex d+ (ReLU(v))
    agree on which elements are activated (>0)."""
    mismatches = 0
    total = n * 5
    boundary_cases = 0
    for _ in range(n):
        answers = [random_v_answer() for _ in range(n_q)]
        p_simplex = score_simplex_p(answers)
        v_simplex = score_simplex_v(answers)
        p_l2 = score_l2_p(answers)

        d_simplex = relu(v_simplex)
        d_l2 = relu(p_to_v_l2(p_l2))

        for i in range(5):
            sim_act = 1 if d_simplex[i] > 1e-12 else 0
            l2_act = 1 if d_l2[i] > 1e-12 else 0
            if sim_act != l2_act:
                mismatches += 1
                # Count boundary cases: element near 0.2 in simplex
                if abs(p_simplex[i] - 0.2) < 0.03:
                    boundary_cases += 1
    return mismatches, boundary_cases, total


# =============================================================================
# Test 3: Classification consistency & distribution
# =============================================================================

def test_classification(n=100000, n_q=32):
    """Compare simplex and L2 classification, check uniformity."""
    sim_counts = Counter()
    l2_counts = Counter()
    agreements = 0
    disagreements = Counter()

    for _ in range(n):
        answers = [random_v_answer() for _ in range(n_q)]
        v_simplex = score_simplex_v(answers)
        p_l2 = score_l2_p(answers)

        label_sim = classify_simplex(v_simplex)
        label_l2 = classify_l2(p_l2)

        sim_counts[label_sim] += 1
        l2_counts[label_l2] += 1

        if label_sim == label_l2:
            agreements += 1
        else:
            pair = f"{label_sim}→{label_l2}"
            disagreements[pair] += 1

    return sim_counts, l2_counts, agreements, disagreements


# =============================================================================
# Test 4: Prototype self-classification
# =============================================================================

def test_prototype_self():
    """Verify all 25 L2 prototypes classify to themselves."""
    failures = []
    for label, proto in PROTOTYPES.items():
        result = classify_l2(proto)
        if result != label:
            failures.append((label, result))
    return failures


# =============================================================================
# Test 5: 顺逆平 distribution
# =============================================================================

def test_synastry_distribution(eps=0.01):
    """Compute 顺/逆/平 across all 300 unordered pairs of 25 types."""
    shun_pairs = []
    ni_pairs = []
    ping_pairs = []

    for i, a in enumerate(ALL_TYPES):
        for j, b in enumerate(ALL_TYPES):
            if i >= j:
                continue
            pA, pB = PROTOTYPES[a], PROTOTYPES[b]
            result, sh, ke = compute_synastry(pA, pB, eps)
            if result == '顺':
                shun_pairs.append((a, b, sh, ke))
            elif result == '逆':
                ni_pairs.append((a, b, sh, ke))
            else:
                ping_pairs.append((a, b, sh, ke))

    # Also compute for same-type self (diagonal)
    self_pairs = []
    for a in ALL_TYPES:
        p = PROTOTYPES[a]
        result, sh, ke = compute_synastry(p, p, eps)
        self_pairs.append((a, result, sh, ke))

    # Pure × pure summary
    pure = ['W', 'F', 'E', 'M', 'R']
    pure_matrix = {}
    for i, a in enumerate(pure):
        for j, b in enumerate(pure):
            if i > j:
                continue
            pA, pB = PROTOTYPES[a], PROTOTYPES[b]
            result, sh, ke = compute_synastry(pA, pB, eps)
            pure_matrix[f"{a}+{b}"] = result

    return shun_pairs, ni_pairs, ping_pairs, self_pairs, pure_matrix


# =============================================================================
# Test 6: Classification distribution uniformity under L2
# =============================================================================

def print_uniformity(counts, n, label="L2"):
    """Print 25-type distribution with uniformity stats."""
    print(f"\n--- {label} 25-Type Distribution (n={n:,}) ---")
    expected = n / 25

    # Pure types
    pure = ['W', 'F', 'E', 'M', 'R']
    composite = ['WF','FW','FE','EF','EM','ME','MR','RM','RW','WR',
                 'WE','EW','FM','MF','ER','RE','MW','WM','RF','FR']

    for group_name, codes in [('Pure', pure), ('Composite', composite)]:
        print(f"  {group_name}:")
        for code in codes:
            c = counts.get(code, 0)
            pct = c / n * 100
            bar_len = int(c / expected * 20) if expected > 0 else 0
            bar = '█' * min(bar_len, 40)
            print(f"    {code}: {c:7d} ({pct:5.2f}%) {bar}")

    counts_list = [counts.get(c, 0) for c in ALL_TYPES]
    min_c = min(counts_list)
    max_c = max(counts_list)
    ratio = max_c / min_c if min_c > 0 else float('inf')
    print(f"\n  Expected per type: {expected:.0f}")
    print(f"  Range: {min_c} - {max_c}  (ratio: {ratio:.2f}x)")

    # Per-element sums (combining all types with same primary)
    elem_sums = defaultdict(int)
    for code, c in counts.items():
        elem_sums[code[0]] += c
    print(f"  Primary element totals:")
    for elem in ['W','F','E','M','R']:
        s = elem_sums[elem]
        pct = s / n * 100
        print(f"    {elem}: {s:7d} ({pct:5.2f}%)")

    return ratio


# =============================================================================
# Test 7: Flow scale comparison
# =============================================================================

def test_flow_scale():
    """Compare seasonal effect magnitude between simplex and L2 models."""
    # Simplex: seasonal effect on v (range ~±1), flowScale=0.4
    # L2: seasonal effect on (p - neutral) (range ~±0.55)
    # Relative strength: effect_mag / max_deviation_mag

    # Simplex max deviation magnitude (W pure)
    v_w = [0.8, -0.2, -0.2, -0.2, -0.2]
    sim_max_dev = sum(abs(x) for x in v_w) / 5  # mean abs deviation

    # L2 max deviation magnitude (W pure)
    p_w = PROTOTYPES['W']
    v_l2_w = p_to_v_l2(p_w)
    l2_max_dev = sum(abs(x) for x in v_l2_w) / 5

    # Seasonal effect (Wood month 寅月)
    s_month = [1, 0, 0, 0, 0]  # Wood month
    gen = mat_vec_mul([[S[j][i] for j in range(5)] for i in range(5)], s_month)  # S^T * s_month
    res = mat_vec_mul([[C[j][i] for j in range(5)] for i in range(5)], s_month)  # C^T * s_month
    effect = [0.4 * (gen[i] - res[i]) for i in range(5)]  # current flowScale=0.4
    effect_mag = sum(abs(x) for x in effect) / 5

    # Effective relative strength
    sim_rel = effect_mag / sim_max_dev
    l2_rel = effect_mag / l2_max_dev

    # Suggested flowScale for L2 to match simplex relative strength
    suggested_scale = 0.4 * (l2_max_dev / sim_max_dev)

    return sim_max_dev, l2_max_dev, effect_mag, sim_rel, l2_rel, suggested_scale


# =============================================================================
# Test 8: Bond output comparison
# =============================================================================

def test_bond_comparison(n=500, n_q=32):
    """Compare Bond Delta vectors between simplex and L2 models."""
    max_delta_diff = 0.0
    max_eff_diff = 0.0

    for _ in range(n):
        answers_a = [random_v_answer() for _ in range(n_q)]
        answers_b = [random_v_answer() for _ in range(n_q)]

        # Simplex
        vA = score_simplex_v(answers_a)
        vB = score_simplex_v(answers_b)
        dA_relu = relu(vA)
        dB_relu = relu(vB)
        deltaA_sim = [sum(S[j][i] * dB_relu[j] for j in range(5)) -
                       sum(C[j][i] * dB_relu[j] for j in range(5)) for i in range(5)]
        effA_sim = [vA[i] + deltaA_sim[i] for i in range(5)]

        # L2
        pA = score_l2_p(answers_a)
        pB = score_l2_p(answers_b)
        dA_l2 = relu(p_to_v_l2(pA))
        dB_l2 = relu(p_to_v_l2(pB))
        deltaA_l2 = [sum(S[j][i] * dB_l2[j] for j in range(5)) -
                      sum(C[j][i] * dB_l2[j] for j in range(5)) for i in range(5)]
        effA_l2 = [p_to_v_l2(pA)[i] + deltaA_l2[i] for i in range(5)]

        # Compare delta direction using cosine similarity
        mag_sim = math.sqrt(sum(x*x for x in deltaA_sim))
        mag_l2 = math.sqrt(sum(x*x for x in deltaA_l2))
        if mag_sim > 1e-12 and mag_l2 > 1e-12:
            cos_sim = dot(deltaA_sim, deltaA_l2) / (mag_sim * mag_l2)
            diff = abs(1.0 - cos_sim)
            max_delta_diff = max(max_delta_diff, diff)

        mag_eff_sim = math.sqrt(sum(x*x for x in effA_sim))
        mag_eff_l2 = math.sqrt(sum(x*x for x in effA_l2))
        if mag_eff_sim > 1e-12 and mag_eff_l2 > 1e-12:
            cos_eff = dot(effA_sim, effA_l2) / (mag_eff_sim * mag_eff_l2)
            diff_eff = abs(1.0 - cos_eff)
            max_eff_diff = max(max_eff_diff, diff_eff)

    return max_delta_diff, max_eff_diff


# =============================================================================
# Main
# =============================================================================

if __name__ == '__main__':
    print("=" * 60)
    print("25types L2 Profile Model — Full Verification")
    print("=" * 60)

    # Config
    N_Q = 32
    N_SIM = 100000
    N_SMALL = 5000

    # --- Test 1: Ordering preservation ---
    print(f"\n{'='*60}")
    print("Test 1: Element Ordering Preservation")
    print("=" * 60)
    swaps, total = test_ordering_preservation(N_SMALL, N_Q)
    print(f"  Pairwise rank swaps (L2 vs simplex): {swaps}/{total} ({swaps/total*100:.4f}%)")
    print(f"  {'PASS: order preserved' if swaps == 0 else 'FAIL: ordering changed'}")

    # --- Test 2: d+ activation equivalence ---
    print(f"\n{'='*60}")
    print("Test 2: d+ Activation Equivalence")
    print("=" * 60)
    mismatches, boundary, total_act = test_dplus_equivalence(N_SMALL, N_Q)
    print(f"  Activation mismatches: {mismatches}/{total_act} ({mismatches/total_act*100:.4f}%)")
    print(f"  Boundary cases (< 3% from threshold): {boundary}/{mismatches if mismatches > 0 else 1}")
    print(f"  {'PASS' if mismatches < total_act * 0.001 else 'ACCEPTABLE (boundary only)' if mismatches == boundary else 'WARNING'}")

    # --- Test 3: Classification ---
    print(f"\n{'='*60}")
    print(f"Test 3: Classification Consistency (n={N_SIM:,})")
    print("=" * 60)
    sim_counts, l2_counts, agreements, disagreements = test_classification(N_SIM, N_Q)
    agree_rate = agreements / N_SIM * 100
    print(f"  Agreement rate: {agreements}/{N_SIM} ({agree_rate:.2f}%)")
    top_disagreements = disagreements.most_common(10)
    if top_disagreements:
        print(f"  Top disagreements:")
        for pair, count in top_disagreements[:5]:
            pct = count / N_SIM * 100
            print(f"    {pair}: {count} ({pct:.3f}%)")
    if agree_rate < 99.9:
        print(f"  NOTE: Disagreements are expected near Voronoi boundaries.")

    # Distribution under both models
    r_sim = print_uniformity(sim_counts, N_SIM, "Simplex")
    r_l2 = print_uniformity(l2_counts, N_SIM, "L2")

    # --- Test 4: Prototype self-classification ---
    print(f"\n{'='*60}")
    print("Test 4: Prototype Self-Classification")
    print("=" * 60)
    failures = test_prototype_self()
    if failures:
        for proto, result in failures:
            print(f"  FAIL: {proto} → {result}")
    else:
        print(f"  All 25 prototypes classify to themselves ✓")

    # --- Test 5: 顺逆平 Distribution ---
    print(f"\n{'='*60}")
    print("Test 5: 顺逆平 (Synastry) Distribution")
    print("=" * 60)
    shun, ni, ping, self_pairs, pure_matrix = test_synastry_distribution(eps=0.01)

    print(f"\n  Pure × Pure (5×5):")
    print(f"      W     F     E     M     R")
    for a in ['W','F','E','M','R']:
        row = f"  {a}  "
        for b in ['W','F','E','M','R']:
            if a == b:
                row += f"   平  "
            else:
                # Build key in insertion order (i <= j from creation loop)
                pure_order = ['W','F','E','M','R']
                ai = pure_order.index(a)
                bi = pure_order.index(b)
                key = f"{a}+{b}" if ai <= bi else f"{b}+{a}"
                val = pure_matrix.get(key, '?')
                row += f"  {val:2s}  "
        print(row)

    print(f"\n  Unordered pairs (300 total, eps=0.01):")
    print(f"    顺: {len(shun):3d}  ({len(shun)/300*100:.1f}%)")
    print(f"    逆: {len(ni):3d}  ({len(ni)/300*100:.1f}%)")
    print(f"    平: {len(ping):3d}  ({len(ping)/300*100:.1f}%)")

    # Show ping pairs
    if ping:
        print(f"\n  平 pairs:")
        for a, b, sh, ke in sorted(ping):
            print(f"    {a}+{b}: sh={sh:.4f} ke={ke:.4f} diff={abs(sh-ke):.4f}")

    # Self-pairs
    print(f"\n  Self-pairs (diagonal):")
    self_shun = [x for x in self_pairs if x[1] == '顺']
    self_ni = [x for x in self_pairs if x[1] == '逆']
    self_ping = [x for x in self_pairs if x[1] == '平']
    print(f"    顺: {len(self_shun)}, 逆: {len(self_ni)}, 平: {len(self_ping)}")
    if self_ping:
        print(f"    平: {', '.join(x[0] for x in self_ping)}")

    # Score distribution
    all_sh_scores = [sh for _, _, sh, _ in shun + ni + ping]
    all_ke_scores = [ke for _, _, _, ke in shun + ni + ping]
    all_diffs = [abs(sh - ke) for _, _, sh, ke in shun + ni + ping]
    print(f"\n  Score statistics:")
    print(f"    shengScore: min={min(all_sh_scores):.4f} max={max(all_sh_scores):.4f} mean={sum(all_sh_scores)/len(all_sh_scores):.4f}")
    print(f"    keScore:    min={min(all_ke_scores):.4f} max={max(all_ke_scores):.4f} mean={sum(all_ke_scores)/len(all_ke_scores):.4f}")
    print(f"    |sh-ke|:    min={min(all_diffs):.4f} max={max(all_diffs):.4f} mean={sum(all_diffs)/len(all_diffs):.4f}")

    # Epsilon sensitivity
    print(f"\n  Epsilon sensitivity (threshold effect):")
    for ep in [0.001, 0.005, 0.01, 0.05, 0.1]:
        s, n, p = 0, 0, 0
        for a in ALL_TYPES:
            for b in ALL_TYPES:
                if a >= b: continue
                pA, pB = PROTOTYPES[a], PROTOTYPES[b]
                sh = dot(pA, mat_vec_mul(SpSt, pB))
                ke = dot(pA, mat_vec_mul(CpCt, pB))
                if sh > ke + ep: s += 1
                elif ke > sh + ep: n += 1
                else: p += 1
        print(f"    eps={ep:.3f}: 顺={s:3d} 逆={n:3d} 平={p:3d}")

    # Confirm: full p vs d+ comparison
    print(f"\n  Full p vs d+ comparison (spot-check):")
    for a, b in [('W', 'F'), ('W', 'M'), ('WF', 'WM'), ('EF', 'MW')]:
        pA, pB = PROTOTYPES[a], PROTOTYPES[b]
        # Full p
        sh_p = dot(pA, mat_vec_mul(SpSt, pB))
        ke_p = dot(pA, mat_vec_mul(CpCt, pB))
        # d+ only
        dA = relu(p_to_v_l2(pA))
        dB = relu(p_to_v_l2(pB))
        sh_d = dot(dA, mat_vec_mul(SpSt, dB))
        ke_d = dot(dA, mat_vec_mul(CpCt, dB))
        res_p = '顺' if sh_p > ke_p+0.01 else ('逆' if ke_p > sh_p+0.01 else '平')
        res_d = '顺' if sh_d > ke_d+0.01 else ('逆' if ke_d > sh_d+0.01 else '平')
        print(f"    {a}+{b}: full_p={res_p}(sh={sh_p:.3f} ke={ke_p:.3f})  d+={res_d}(sh={sh_d:.3f} ke={ke_d:.3f}) {'SAME' if res_p == res_d else 'DIFFER'}")

    # Element × element analysis for pure types
    print(f"\n  Pure × Pure 顺逆 pattern (off-diagonal):")
    pure_order = ['W','F','E','M','R']
    for i, a in enumerate(pure_order):
        counts = {'顺': 0, '逆': 0, '平': 0}
        for j, b in enumerate(pure_order):
            if i == j: continue
            if i < j:
                key = f"{a}+{b}"
            else:
                key = f"{b}+{a}"
            val = pure_matrix.get(key, '?')
            if val in counts:
                counts[val] += 1
        print(f"    {a}: {counts['顺']}顺 {counts['逆']}逆 {counts['平']}平")

    # --- Test 6: Flow scale ---
    print(f"\n{'='*60}")
    print("Test 6: Flow Scale Comparison")
    print("=" * 60)
    sim_dev, l2_dev, eff_mag, sim_rel, l2_rel, sugg = test_flow_scale()
    print(f"  Max mean |deviation| — simplex: {sim_dev:.4f}, L2: {l2_dev:.4f}")
    print(f"  Seasonal effect mean |magnitude|: {eff_mag:.4f}")
    print(f"  Relative strength — simplex: {sim_rel:.2%}, L2: {l2_rel:.2%}")
    print(f"  Suggested flowScale for L2: {sugg:.3f} (current=0.4)")

    # --- Test 7: Bond consistency ---
    print(f"\n{'='*60}")
    print("Test 7: Bond Delta Direction Consistency")
    print("=" * 60)
    max_d, max_e = test_bond_comparison(500, N_Q)
    print(f"  Max |1 - cos(Δ_simplex, Δ_L2)|: {max_d:.6f}")
    print(f"  Max |1 - cos(eff_simplex, eff_L2)|: {max_e:.6f}")
    ok = max_d < 0.01 and max_e < 0.01
    print(f"  {'PASS: Bond direction preserved' if ok else 'WARNING: Bond direction shifted'}")

    # --- Summary ---
    print(f"\n{'='*60}")
    print("Summary")
    print("=" * 60)
    print(f"  ✓ Element ordering: {'PRESERVED' if swaps == 0 else 'CHANGED'}")
    print(f"  ✓ d+ activation: {'EQUIVALENT' if mismatches == boundary else 'BOUNDARY ONLY'}")
    print(f"  ✓ Classification agreement: {agree_rate:.2f}%")
    print(f"  ✓ Simplex uniformity ratio: {r_sim:.2f}x")
    print(f"  ✓ L2 uniformity ratio: {r_l2:.2f}x")
    print(f"  ✓ Prototype self-classification: {'ALL PASS' if not failures else 'FAILURES'}")
    print(f"  ✓ 顺逆平: {len(shun)}/{len(ni)}/{len(ping)} (300 pairs)")
    print(f"  ✓ Bond direction: max diff {max_d:.6f}")
    print(f"  ✓ Suggested flowScale: {sugg:.3f}")
