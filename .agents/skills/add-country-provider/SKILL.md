# Skill: Add Country & Provider (SDD-Compatible)

This skill automates the expansion of GlobalTask Bank to new countries and bank providers. It follows Spec-Driven Development (SDD) principles by first gathering requirements and then executing the multi-layer implementation.

## Workflow

1.  **Gather Requirements**: Ask the user for:
    *   Country Name and ISO Code (e.g., Chile, CL).
    *   Currency (e.g., CLP).
    *   Bank Providers (Names and if they provide verified income).
    *   Risk Rules:
        *   Maximum monthly payment percentage (e.g., 30%).
        *   Maximum amount multiplier (e.g., 5x income).
        *   Special rules (e.g., rejected if requested amount > 200% income).

2.  **Plan Changes**:
    *   **Database**: Migration for `countries` and `bank_providers` tables.
    *   **Backend Config**: Update `backend/config/workflows.yaml` with the new ISO code and risk rules.
    *   **Backend Logic**: Adjust `internal/application/service/loan_application.go` if custom income parsing is needed.
    *   **Mock Bank**: Update `backend/cmd/mockbank/main.go` to simulate the new provider's behavior.
    *   **Frontend**: Update `DashboardView.vue` (flags, dropdowns) and `loanApplications.js` (country mapping).

3.  **Implementation**:
    *   Execute the SQL migration in the live DB and save to files.
    *   Update YAML configurations.
    *   Apply code changes in Go and Vue.
    *   Verify with a test application.

## Guidelines

*   **Flags**: Use standard emojis for country flags in the frontend.
*   **IDs**: Follow the incremental sequence for country IDs.
*   **Atomic**: Ensure DB and config changes are applied together to avoid inconsistencies.
*   **Validation**: Always validate that the ISO code is unique.

## Trigger Phrase
"I want to add a new country" or "Add [Country] to the bank" or "New provider for [Country]".
