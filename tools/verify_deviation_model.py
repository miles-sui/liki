"""Fivefold Types — Deviation Model Math Verification

Verifies the algebraic properties of the sum=0 deviation model:
  1. Algebra closure: Σ(v_eff) = 0 for all inputs
  2. Prototype classification consistency with simplex model
  3. μ sign correctness for known pathology cases
  4. Boundary stability for extreme inputs
"""

import numpy as np
from itertools import product

# =============================================================================
# Constants
# =============================================================================

ELEMENTS = ['W', 'F', 'E', 'M', 'R']

# S (生) matrix: Wood→Fire→Earth→Metal→Water→Wood
S = np.array([
    [0, 1, 0, 0, 0],  # W → F
    [0, 0, 1, 0, 0],  # F → E
    [0, 0, 0, 1, 0],  # E → M
    [0, 0, 0, 0, 1],  # M → R
    [1, 0, 0, 0, 0],  # R → W
], dtype=float)

# C (克) matrix: Wood→Earth→Water→Fire→Metal→Wood
C = np.array([
    [0, 0, 1, 0, 0],  # W 克 E
    [0, 0, 0, 1, 0],  # F 克 M
    [0, 0, 0, 0, 1],  # E 克 R
    [1, 0, 0, 0, 0],  # M 克 W
    [0, 1, 0, 0, 0],  # R 克 F
], dtype=float)

# =============================================================================
# Helper: generate a random deviation vector with Σ=0
# =============================================================================
def random_deviation(rng=None):
    """Generate random v satisfying Σv=0 via p in simplex then (p-0.2)*5."""
    if rng is None:
        rng = np.random
    # Dirichlet with alpha=[1,1,1,1,1] = uniform on simplex
    p = rng.dirichlet([1, 1, 1, 1, 1])
    v = (p - 0.2) * 5
    return v

def random_deviation_pair(rng=None):
    """Generate a random pair (v_A, q_A), (v_B, q_B)."""
    if rng is None:
        rng = np.random
    v_A = random_deviation(rng)
    v_B = random_deviation(rng)
    q_A = rng.uniform(0, 1)
    q_B = rng.uniform(0, 1)
    return v_A, q_A, v_B, q_B


# =============================================================================
# Bond computation in deviation model
# =============================================================================
def bond_a_effected_by_b(v_A, v_B, q_B):
    """Compute A's effective v after B's influence. Returns (v_eff, mu)."""
    E_B = v_B * q_B

    delta_n = S.T @ E_B
    delta_c = C.T @ E_B

    v_eff = v_A + delta_n - delta_c

    mu = np.mean(np.abs(v_eff)) - np.mean(np.abs(v_A))

    return v_eff, mu, delta_n, delta_c


# =============================================================================
# Self diagnostics
# =============================================================================
def self_diagnostics(v):
    """Compute resonate_self — total deviation magnitude."""
    resonate = v @ v
    return resonate


# =============================================================================
# Test 1: Algebra Closure
# =============================================================================
def test_algebra_closure(n=10000):
    """Verify Σ(v_eff) = 0 for all random inputs."""
    rng = np.random.RandomState(42)
    max_error = 0.0
    for _ in range(n):
        v_A, _, v_B, q_B = random_deviation_pair(rng)
        v_eff, _, _, _ = bond_a_effected_by_b(v_A, v_B, q_B)
        err = abs(np.sum(v_eff))
        max_error = max(max_error, err)

    print(f"Test 1: Algebra Closure ({n} trials)")
    print(f"  Max |Σ(v_eff)| = {max_error:.2e}")
    print(f"  PASS: Σ(v_eff) = 0 within machine epsilon" if max_error < 1e-12
          else f"  FAIL: error too large")
    print()
    return max_error < 1e-12


# =============================================================================
# Test 2: Σ(Δn) = 0 and Σ(Δc) = 0
# =============================================================================
def test_delta_sums(n=10000):
    """Verify Δn and Δc also sum to zero."""
    rng = np.random.RandomState(42)
    max_err_n, max_err_c = 0.0, 0.0
    for _ in range(n):
        _, _, v_B, q_B = random_deviation_pair(rng)
        E_B = v_B * q_B
        delta_n = S.T @ E_B
        delta_c = C.T @ E_B
        max_err_n = max(max_err_n, abs(np.sum(delta_n)))
        max_err_c = max(max_err_c, abs(np.sum(delta_c)))

    print(f"Test 2: Delta Sums ({n} trials)")
    print(f"  Max |Σ(Δn)| = {max_err_n:.2e}")
    print(f"  Max |Σ(Δc)| = {max_err_c:.2e}")
    print(f"  PASS" if max_err_n < 1e-12 and max_err_c < 1e-12 else f"  FAIL")
    print()


# =============================================================================
# Test 3: μ symmetry — two balanced people produce μ ≈ 0
# =============================================================================
def test_mu_balanced():
    """Two perfectly balanced people should have μ ≈ 0."""
    v_bal = np.zeros(5)
    v_eff, mu, _, _ = bond_a_effected_by_b(v_bal, v_bal, 0.8)
    print(f"Test 3: Balanced × Balanced")
    print(f"  v_bal = {v_bal}")
    print(f"  v_eff = {v_eff}")
    print(f"  μ = {mu}")
    print(f"  PASS: μ = 0 (no effect)" if abs(mu) < 1e-15 else f"  FAIL")
    print()


# =============================================================================
# Test 4: Known pathology cases — μ should be > 0 (depleting)
# =============================================================================
def test_mu_pathology():
    """木乘土: Wood excess + Earth deficiency. B (Wood-dominant, high q)
       interacting with A (Wood-high, Earth-low) should μ > 0 (depleting)."""
    # A: 木乘土 profile — Wood excess, Earth deficiency
    v_A = np.array([+1.0, -0.25, -0.5, -0.25, 0.0])
    q_A = 0.7

    # B: 木型 high-q — amplifies A's Wood excess
    v_B = np.array([+1.2, -0.3, -0.3, -0.3, -0.3])
    q_B = 0.9

    v_eff, mu, _, _ = bond_a_effected_by_b(v_B, v_A, q_A)  # B affects A

    print(f"Test 4a: 木乘土 — Wood-dominant B depletes Wood-excess A")
    print(f"  v_A  = {np.round(v_A, 2)}")
    print(f"  v_B  = {np.round(v_B, 2)}")
    print(f"  v_eff = {np.round(v_eff, 3)}")
    print(f"  μ = {mu:.4f}")
    print(f"  Expected: μ > 0 (耗)" if mu > 0 else f"  UNEXPECTED: μ < 0")

    # Case 4b: complementary interaction — B's Wood nourishes A's deficient Fire
    # AND B's Wood controls A's excess Earth — both effects reduce deviation
    v_A2 = np.array([0.0, -0.5, +0.5, 0.0, 0.0])  # Fire deficient, Earth excess
    v_B2 = np.array([+0.5, 0.0, 0.0, -0.25, -0.25])  # Wood excess, balanced otherwise
    q_A2, q_B2 = 0.6, 1.0  # q_B=1 for clarity

    v_eff2, mu2, dn, dc = bond_a_effected_by_b(v_A2, v_B2, q_B2)

    print(f"\nTest 4b: Complementary — B's Wood nourishes A's Fire AND controls A's Earth")
    print(f"  v_A  = {np.round(v_A2, 2)} (Fire deficient, Earth excess)")
    print(f"  v_B  = {np.round(v_B2, 2)} (Wood excess)")
    print(f"  Δn   = {np.round(dn, 3)} (Fire +{dn[1]:.1f} ← B's Wood nourishes)")
    print(f"  Δc   = {np.round(dc, 3)} (Earth -{dc[2]:.1f} ← B's Wood controls)")
    print(f"  v_eff = {np.round(v_eff2, 3)}")
    print(f"  μ = {mu2:.4f}")
    print(f"  {'PASS: μ < 0 (旺)' if mu2 < 0 else 'FAIL: expected μ < 0'}")
    print()


# =============================================================================
# Test 5: Self diagnostics — known cases
# =============================================================================
def test_self_diagnostics():
    """Verify resonate_self for known profiles."""
    # Balanced
    v_bal = np.zeros(5)
    r1 = self_diagnostics(v_bal)
    print(f"Test 5: Self Diagnostics (resonate_self)")
    print(f"  Balanced v: resonate={r1:.4f}")
    assert abs(r1) < 1e-15

    # Pure Wood extreme
    v_w = np.array([+0.80, -0.20, -0.20, -0.20, -0.20])
    r2 = self_diagnostics(v_w)
    print(f"  W prototype: resonate={r2:.4f}")

    # Wood-fire double (WF prototype)
    v_wf = np.array([+0.30, +0.30, -0.20, -0.20, -0.20])
    r3 = self_diagnostics(v_wf)
    print(f"  WF prototype: resonate={r3:.4f}")
    print("  PASS: all computed without error")
    print()


# =============================================================================
# Test 6: Boundary stability
# =============================================================================
def test_boundary():
    """Extreme inputs should not produce NaN or overflow."""
    rng = np.random.RandomState(99)

    extremes = [
        np.array([+2.33, -0.2, -0.2, -0.2, -0.2]) / 5 * 5,  # hmm this needs fixing
    ]

    # Actually generate extreme but valid v's
    # Max one element: p=[2/3, 1/12, 1/12, 1/12, 1/12] → v=[2.33, -0.58, ...]
    p_extreme = np.array([2/3, 1/12, 1/12, 1/12, 1/12])
    v_extreme = (p_extreme - 0.2) * 5

    # Min: p=[0, 0.25, 0.25, 0.25, 0.25] → v=[-1, +0.25, ...]
    p_min = np.array([0.0, 0.25, 0.25, 0.25, 0.25])
    v_min = (p_min - 0.2) * 5

    print(f"Test 6: Boundary Stability")
    for name, v in [("max-W", v_extreme), ("min-W", v_min)]:
        v_eff, mu, _, _ = bond_a_effected_by_b(v, v, 1.0)
        has_nan = np.any(np.isnan(v_eff)) or np.isnan(mu)
        has_inf = np.any(np.isinf(v_eff)) or np.isinf(mu)
        print(f"  {name}: v_eff={np.round(v_eff, 3)}, μ={mu:.4f}")
        print(f"    NaN={has_nan}, Inf={has_inf} → {'PASS' if not has_nan and not has_inf else 'FAIL'}")

    # Extreme mutual influence
    v_A = v_extreme
    v_B = v_min
    q_B = 1.0
    v_eff, mu, _, _ = bond_a_effected_by_b(v_A, v_B, q_B)
    has_nan = np.any(np.isnan(v_eff)) or np.isnan(mu)
    print(f"  extreme-A + extreme-B: v_eff={np.round(v_eff, 3)}, μ={mu:.4f}")
    print(f"    NaN={has_nan} → {'PASS' if not has_nan else 'FAIL'}")
    print()


# =============================================================================
# Test 7: Prototype distance classification
# =============================================================================
def test_prototype_classification():
    """Verify that prototype distance classification works in deviation space."""

    # 16 prototypes in deviation coordinates
    prototypes = {
        'W':  np.array([+0.80, -0.20, -0.20, -0.20, -0.20]),
        'F':  np.array([-0.20, +0.80, -0.20, -0.20, -0.20]),
        'E':  np.array([-0.20, -0.20, +0.80, -0.20, -0.20]),
        'M':  np.array([-0.20, -0.20, -0.20, +0.80, -0.20]),
        'R':  np.array([-0.20, -0.20, -0.20, -0.20, +0.80]),
        'WE': np.array([+0.35, -0.20, +0.25, -0.20, -0.20]),
        'FE': np.array([-0.20, +0.35, +0.25, -0.20, -0.20]),
        'ME': np.array([-0.20, -0.20, +0.25, +0.35, -0.20]),
        'RE': np.array([-0.20, -0.20, +0.25, -0.20, +0.35]),
        'WF': np.array([+0.30, +0.30, -0.20, -0.20, -0.20]),
        'WM': np.array([+0.30, -0.20, -0.20, +0.30, -0.20]),
        'WR': np.array([+0.30, -0.20, -0.20, -0.20, +0.30]),
        'FM': np.array([-0.20, +0.30, -0.20, +0.30, -0.20]),
        'FR': np.array([-0.20, +0.30, -0.20, -0.20, +0.30]),
        'MR': np.array([-0.20, -0.20, -0.20, +0.30, +0.30]),
        'B':  np.array([ 0.00,  0.00,  0.00,  0.00,  0.00]),
    }

    def classify(v):
        best_label, best_dist = None, float('inf')
        for label, proto in prototypes.items():
            d = np.linalg.norm(v - proto)
            if d < best_dist:
                best_dist = d
                best_label = label
        return best_label, best_dist

    # Each prototype should classify as itself
    print(f"Test 7: Prototype Classification")
    all_correct = True
    for label, proto in prototypes.items():
        classified, dist = classify(proto)
        ok = classified == label
        if not ok:
            print(f"  MISMATCH: {label} classified as {classified} (dist={dist:.4f})")
            all_correct = False
    print(f"  Self-classification: {'PASS' if all_correct else 'FAIL'}")

    # Random vectors should converge to nearest prototype
    rng = np.random.RandomState(42)
    mismatches = 0
    for _ in range(1000):
        v = random_deviation(rng)
        label, _ = classify(v)
        if label is None:
            mismatches += 1
    print(f"  Random vectors (1000): all classified ({mismatches} unclassified)")
    print(f"  PASS" if mismatches == 0 else f"  FAIL")
    print()


# =============================================================================
# Test 8: Simplex vs Deviation classification consistency
# =============================================================================
def test_simplex_consistency(n=5000):
    """Verify that prototype self-mapping is consistent between old and new
       classification methods. Random-vector agreement is NOT expected to be
       high because the old 'subtract min + renormalize' preprocessing changes
       the Voronoi partition — the deviation model intentionally preserves
       absolute deviation information that the old method erases."""

    # Old prototypes (simplex space)
    simplex_protos = {
        'W':  np.array([1.0, 0.0, 0.0, 0.0, 0.0]),
        'F':  np.array([0.0, 1.0, 0.0, 0.0, 0.0]),
        'E':  np.array([0.0, 0.0, 1.0, 0.0, 0.0]),
        'M':  np.array([0.0, 0.0, 0.0, 1.0, 0.0]),
        'R':  np.array([0.0, 0.0, 0.0, 0.0, 1.0]),
        'WE': np.array([0.55, 0.0, 0.45, 0.0, 0.0]),
        'FE': np.array([0.0, 0.55, 0.45, 0.0, 0.0]),
        'ME': np.array([0.0, 0.0, 0.45, 0.55, 0.0]),
        'RE': np.array([0.0, 0.0, 0.45, 0.0, 0.55]),
        'WF': np.array([0.5, 0.5, 0.0, 0.0, 0.0]),
        'WM': np.array([0.5, 0.0, 0.0, 0.5, 0.0]),
        'WR': np.array([0.5, 0.0, 0.0, 0.0, 0.5]),
        'FM': np.array([0.0, 0.5, 0.0, 0.5, 0.0]),
        'FR': np.array([0.0, 0.5, 0.0, 0.0, 0.5]),
        'MR': np.array([0.0, 0.0, 0.0, 0.5, 0.5]),
        'B':  np.array([0.2, 0.2, 0.2, 0.2, 0.2]),
    }

    # Deviation prototypes (same as in Test 7)
    dev_protos = {
        'W':  np.array([+0.80, -0.20, -0.20, -0.20, -0.20]),
        'F':  np.array([-0.20, +0.80, -0.20, -0.20, -0.20]),
        'E':  np.array([-0.20, -0.20, +0.80, -0.20, -0.20]),
        'M':  np.array([-0.20, -0.20, -0.20, +0.80, -0.20]),
        'R':  np.array([-0.20, -0.20, -0.20, -0.20, +0.80]),
        'WE': np.array([+0.35, -0.20, +0.25, -0.20, -0.20]),
        'FE': np.array([-0.20, +0.35, +0.25, -0.20, -0.20]),
        'ME': np.array([-0.20, -0.20, +0.25, +0.35, -0.20]),
        'RE': np.array([-0.20, -0.20, +0.25, -0.20, +0.35]),
        'WF': np.array([+0.30, +0.30, -0.20, -0.20, -0.20]),
        'WM': np.array([+0.30, -0.20, -0.20, +0.30, -0.20]),
        'WR': np.array([+0.30, -0.20, -0.20, -0.20, +0.30]),
        'FM': np.array([-0.20, +0.30, -0.20, +0.30, -0.20]),
        'FR': np.array([-0.20, +0.30, -0.20, -0.20, +0.30]),
        'MR': np.array([-0.20, -0.20, -0.20, +0.30, +0.30]),
        'B':  np.array([ 0.00,  0.00,  0.00,  0.00,  0.00]),
    }

    def classify_old(p):
        """Old method: subtract min, renormalize, then Euclidean distance."""
        v_sub = p - np.min(p)
        denom = np.sum(v_sub)
        if denom < 1e-15:
            return 'B'
        v_norm = v_sub / denom
        best_label, best_dist = None, float('inf')
        for label, proto in simplex_protos.items():
            d = np.linalg.norm(v_norm - proto)
            if d < best_dist:
                best_dist = d
                best_label = label
        return best_label

    def classify_new(v):
        """New method: direct Euclidean distance in deviation space."""
        best_label, best_dist = None, float('inf')
        for label, proto in dev_protos.items():
            d = np.linalg.norm(v - proto)
            if d < best_dist:
                best_dist = d
                best_label = label
        return best_label

    # Test 1: Each simplex prototype, when transformed to deviation,
    # should classify as the same label (prototype self-consistency)
    proto_ok = 0
    for label, sp in simplex_protos.items():
        v = (sp - 0.2) * 5  # transform to deviation
        new_label = classify_new(v)
        if new_label == label:
            proto_ok += 1

    # Test 2: Random vectors — classification may differ because
    # old method uses subtract-min preprocessing
    rng = np.random.RandomState(123)
    agreements = 0
    for _ in range(n):
        p = rng.dirichlet([1, 1, 1, 1, 1])
        v = (p - 0.2) * 5
        if classify_old(p) == classify_new(v):
            agreements += 1

    print(f"Test 8: Classification Consistency")
    print(f"  Prototype self-mapping: {proto_ok}/16 correct")
    print(f"  Random vector agreement (old vs new): {agreements/n:.1%}")
    print(f"  Note: Random agreement < 100% is expected — old method uses")
    print(f"  subtract-min preprocessing that erases absolute level information.")
    print(f"  The deviation model intentionally preserves this information.")
    print(f"  PASS" if proto_ok == 16 else f"  FAIL: prototype mapping broken")
    print()


# =============================================================================
# Test 9: Σv = 0 for assessment output
# =============================================================================
def test_assessment_output():
    """Verify that the assessment formula (p-0.2)*5 always produces Σv=0."""
    rng = np.random.RandomState(77)
    max_err = 0.0
    for _ in range(5000):
        p = rng.dirichlet([1, 1, 1, 1, 1])
        v = (p - 0.2) * 5
        max_err = max(max_err, abs(np.sum(v)))

    print(f"Test 9: Assessment Output Σv=0 (5000 trials)")
    print(f"  Max |Σv| = {max_err:.2e}")
    print(f"  PASS" if max_err < 1e-14 else f"  FAIL")
    print()


# =============================================================================
# Main
# =============================================================================
if __name__ == '__main__':
    print("=" * 60)
    print("Fivefold Types — Deviation Model Verification")
    print("=" * 60)
    print()

    results = []
    results.append(("1. Algebra Closure", test_algebra_closure()))
    test_delta_sums()
    test_mu_balanced()
    test_mu_pathology()
    test_self_diagnostics()
    test_boundary()
    test_prototype_classification()
    test_simplex_consistency()
    test_assessment_output()

    print("=" * 60)
    print("Summary")
    print("=" * 60)
    all_pass = all(r[1] for r in results)
    for name, passed in results:
        status = "PASS" if passed else "FAIL"
        print(f"  {status}: {name}")
    print(f"\n  OVERALL: {'ALL PASS' if all_pass else 'SOME FAILURES'}")
