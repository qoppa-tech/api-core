-- name: GetOnboardingSteps :many
SELECT id FROM onboarding_steps
ORDER BY id;

-- name: GetOnboardingOptionsByStep :many
SELECT id, step_id, option_index, created_at FROM onboarding_options
WHERE step_id = $1
ORDER BY option_index;

-- name: GetAllOnboardingOptions :many
SELECT id, step_id, option_index, created_at FROM onboarding_options
ORDER BY step_id, option_index;

-- name: GetUserOnboardingSelections :many
SELECT * FROM salon_onboarding_selection
WHERE user_id = $1
ORDER BY step_id, option_index;

-- name: SaveOnboardingSelection :exec
INSERT INTO salon_onboarding_selection (
    user_id,
    salon_id,
    step_id,
    option_index
) VALUES (
    $1, $2, $3, $4
) ON CONFLICT (salon_id, step_id, option_index) DO NOTHING;

-- name: DeleteUserOnboardingSelections :exec
DELETE FROM salon_onboarding_selection
WHERE user_id = $1 AND step_id = $2;

-- name: UpdateUserOnboardingStep :exec
UPDATE users
SET current_onboarding_step_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteUserOnboarding :exec
UPDATE users
SET current_onboarding_step_id = $2,
    onboarding_completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;
