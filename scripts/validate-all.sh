#!/usr/bin/env bash
#
# validate-all.sh - Validate all formulas and casks in the repository
#
# Usage: ./validate-all.sh
# This script runs brew audit and brew style on all formulas and casks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
error() {
    echo -e "${RED}Error: $1${NC}" >&2
}

success() {
    echo -e "${GREEN}✓ $1${NC}"
}

info() {
    echo -e "${BLUE}→ $1${NC}"
}

warn() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

# Check if brew is available
if ! command -v brew &> /dev/null; then
    error "Homebrew (brew) is not installed. Install it from https://brew.sh"
    exit 1
fi

# Navigate to repository root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$REPO_ROOT"

info "Validating packages in $REPO_ROOT"
echo ""

# Initialize counters
TOTAL_FORMULAS=0
PASSED_FORMULAS=0
FAILED_FORMULAS=0
TOTAL_CASKS=0
PASSED_CASKS=0
FAILED_CASKS=0

# Array to store failed items
declare -a FAILED_ITEMS=()

# Validate formulas
info "Checking for formulas in Formula/..."
if [ -d "Formula" ]; then
    FORMULAS=(Formula/*.rb)
    if [ -e "${FORMULAS[0]}" ]; then
        TOTAL_FORMULAS=${#FORMULAS[@]}
        info "Found $TOTAL_FORMULAS formula(s)"
        echo ""
        
        for formula in "${FORMULAS[@]}"; do
            if [ "$(basename "$formula")" = ".gitkeep" ] || [ ! -f "$formula" ]; then
                continue
            fi
            
            FORMULA_NAME=$(basename "$formula" .rb)
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            info "Validating formula: $FORMULA_NAME"
            echo ""
            
            FORMULA_PASSED=true
            
            # Run brew audit
            info "Running brew audit --strict --online..."
            if brew audit --strict --online "$formula" 2>&1; then
                success "Audit passed for $FORMULA_NAME"
            else
                error "Audit failed for $FORMULA_NAME"
                FAILED_ITEMS+=("Formula: $FORMULA_NAME (audit)")
                ((FAILED_FORMULAS++))
                FORMULA_PASSED=false
            fi
            echo ""
            
            # Run brew style
            info "Running brew style..."
            if brew style "$formula" 2>&1; then
                success "Style check passed for $FORMULA_NAME"
            else
                error "Style check failed for $FORMULA_NAME"
                FAILED_ITEMS+=("Formula: $FORMULA_NAME (style)")
                ((FAILED_FORMULAS++))
                FORMULA_PASSED=false
            fi
            
            # Increment passed counter if all checks passed
            if [ "$FORMULA_PASSED" = true ]; then
                ((PASSED_FORMULAS++))
            fi
            echo ""
        done
    else
        warn "No formulas found in Formula/"
    fi
else
    warn "Formula directory does not exist"
fi

# Validate casks
info "Checking for casks in Casks/..."
if [ -d "Casks" ]; then
    CASKS=(Casks/*.rb)
    if [ -e "${CASKS[0]}" ]; then
        TOTAL_CASKS=${#CASKS[@]}
        info "Found $TOTAL_CASKS cask(s)"
        echo ""
        
        for cask in "${CASKS[@]}"; do
            if [ "$(basename "$cask")" = ".gitkeep" ] || [ ! -f "$cask" ]; then
                continue
            fi
            
            CASK_NAME=$(basename "$cask" .rb)
            echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
            info "Validating cask: $CASK_NAME"
            echo ""
            
            CASK_PASSED=true
            
            # Run brew audit (casks)
            info "Running brew audit --strict --online..."
            if brew audit --cask --strict --online "$cask" 2>&1; then
                success "Audit passed for $CASK_NAME"
            else
                error "Audit failed for $CASK_NAME"
                FAILED_ITEMS+=("Cask: $CASK_NAME (audit)")
                ((FAILED_CASKS++))
                CASK_PASSED=false
            fi
            echo ""
            
            # Run brew style
            info "Running brew style..."
            if brew style "$cask" 2>&1; then
                success "Style check passed for $CASK_NAME"
            else
                error "Style check failed for $CASK_NAME"
                FAILED_ITEMS+=("Cask: $CASK_NAME (style)")
                ((FAILED_CASKS++))
                CASK_PASSED=false
            fi
            
            # Increment passed counter if all checks passed
            if [ "$CASK_PASSED" = true ]; then
                ((PASSED_CASKS++))
            fi
            echo ""
        done
    else
        warn "No casks found in Casks/"
    fi
else
    warn "Casks directory does not exist"
fi

# Display summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${BLUE}Validation Summary${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# Formulas summary
if [ $TOTAL_FORMULAS -gt 0 ]; then
    echo -e "${BLUE}Formulas:${NC}"
    echo "  Total:  $TOTAL_FORMULAS"
    echo -e "  ${GREEN}Passed: $PASSED_FORMULAS${NC}"
    if [ $FAILED_FORMULAS -gt 0 ]; then
        echo -e "  ${RED}Failed: $FAILED_FORMULAS${NC}"
    else
        echo -e "  ${GREEN}Failed: 0${NC}"
    fi
    echo ""
fi

# Casks summary
if [ $TOTAL_CASKS -gt 0 ]; then
    echo -e "${BLUE}Casks:${NC}"
    echo "  Total:  $TOTAL_CASKS"
    echo -e "  ${GREEN}Passed: $PASSED_CASKS${NC}"
    if [ $FAILED_CASKS -gt 0 ]; then
        echo -e "  ${RED}Failed: $FAILED_CASKS${NC}"
    else
        echo -e "  ${GREEN}Failed: 0${NC}"
    fi
    echo ""
fi

# List failed items
if [ ${#FAILED_ITEMS[@]} -gt 0 ]; then
    echo -e "${RED}Failed Items:${NC}"
    for item in "${FAILED_ITEMS[@]}"; do
        echo -e "  ${RED}✗${NC} $item"
    done
    echo ""
fi

# Overall result
TOTAL_ITEMS=$((TOTAL_FORMULAS + TOTAL_CASKS))
TOTAL_FAILURES=$((FAILED_FORMULAS + FAILED_CASKS))

if [ $TOTAL_ITEMS -eq 0 ]; then
    warn "No formulas or casks found to validate"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
fi

if [ $TOTAL_FAILURES -eq 0 ]; then
    echo -e "${GREEN}✓ All validation checks passed!${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 0
else
    echo -e "${RED}✗ Validation failed with $TOTAL_FAILURES error(s)${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    exit 1
fi
