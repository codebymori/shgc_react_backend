#!/bin/bash

# Configuration
BASE_URL="http://127.0.0.1:8000"
API_ENDPOINT="/api/admin/holes"
FULL_URL="${BASE_URL}${API_ENDPOINT}"

# Token (Verified Working)
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiY21pdHN5emY1MDAwMGFiOWZkaWd0dDZvaCIsImVtYWlsIjoiYWRtaW5Ac2VudHVsZ29sZi5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3NjY5MTgwNzUsImlhdCI6MTc2NjgzMTY3NX0.s8QIlLMFOwWeYyg6_ERLl_uXZU33W75P0o2ZWTnF8c0"

echo "Sanity Check (Connectivity)..."
curl -I "${BASE_URL}/api/holes" || echo "Sanity Check FAILED"

echo "Starting to seed 18 holes..."

# Define Hole Data function
create_hole() {
    local id=$1
    local name=$2
    local par=$3
    local distance=$4
    local desc=$5
    local img_path="temp/hole${id}.webp"

    # Extract clean digits from Par (e.g. "Par 4" -> 4)
    local par_val=$(echo $par | tr -dc '0-9')
    
    echo "Creating Hole $id ($name)..."
    
    # Check if image exists
    if [ ! -f "$img_path" ]; then
        echo "Warning: Image $img_path not found! Creating dummy..."
        touch "$img_path"
    fi

    HTTP_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$FULL_URL" \
        -H "Authorization: Bearer $TOKEN" \
        -F "name=$name" \
        -F "par=$par_val" \
        -F "distance=$distance" \
        -F "description=$desc" \
        -F "image=@$img_path")
        
    HTTP_BODY=$(echo "$HTTP_RESPONSE" | head -n -1)
    HTTP_CODE=$(echo "$HTTP_RESPONSE" | tail -n 1)

    if [ "$HTTP_CODE" -eq 201 ] || [ "$HTTP_CODE" -eq 200 ]; then
        echo "Success! ($HTTP_CODE)"
    else
        echo "FAILED to create Hole $id. HTTP Code: $HTTP_CODE"
        echo "Response: $HTTP_BODY"
    fi
}

# --- HOLE DATA ---

create_hole 1 "Hole 1" "Par 4" "383" "A scenic opener across a wide, right-to-left sloping landscape. Navigate the right-side bunker to reach a beautiful green framed by lush trees, though its uneven surface requires a careful touch."
create_hole 2 "Hole 2" "Par 4" "364" "A downhill drive where accuracy is more valuable than distance. Stay center-right to avoid the narrow rough and the water hazard bordering the right side."
create_hole 3 "Hole 3" "Par 5" "438" "Experience a dramatic transition from an uphill start to a downhill finish. The green is uniquely split by a cart path and features a challenging, uneven slope."
create_hole 4 "Hole 4" "Par 3" "108" "Club selection is vital as wind funnels through this valley. Aim for the center of the raised green to ensure a safe landing and avoid difficult recoveries."
create_hole 5 "Hole 5" "Par 4" "364" "A high-concentration hole featuring a narrow fairway and a blind shot. Drive from the heights toward the right side to reach a tiered green guarded by a strategic bunker."
create_hole 6 "Hole 6" "Par 5" "445" "A true test of skill on a hilly fairway, requiring a precise blind tee shot to navigate the rolling terrain."
create_hole 7 "Hole 7" "Par 3" "135" "Distance control is essential to avoid the pond tucked behind the green. Flanked by bunkers, this scenic hole is best played toward the center or front."
create_hole 8 "Hole 8" "Par 4" "343" "Launch from a low-elevation tee up a steep, dogleg fairway. Accurate carry distance is key to clearing the massive front-left bunker and avoiding the hidden trap behind the green"
create_hole 9 "Hole 9" "Par 4" "344" "A rewarding scoring opportunity with a wide landing area. Long hitters can clear the bunker for a very short approach to a friendly, inviting green."
create_hole 10 "Hole 10" "Par 4" "359" "Home to the iconic “Par Tee Time” signage. This downhill hole offers a breathtaking view over a wide fairway dotted with natural rocks and bunkers."
create_hole 11 "Hole 11" "Par 5" "445" "A ruggedly beautiful and challenging hole featuring natural rock formations along the fairway, leading to an expansive, wide green."
create_hole 12 "Hole 12" "Par 3" "126" "The crown jewel of SHL. A breathtakingly beautiful Par 3 framed by boulders and bunkers; a straight tee shot here is rewarded with the course's most iconic view."
create_hole 13 "Hole 13" "Par 4" "262" "A short, hilly Par 4 with a blind tee shot that hides the green from view. The fairway slopes sharply left, leading to a very tricky, sloped green."
create_hole 14 "Hole 14" "Par 4" "334" "A tricky dogleg right set against a backdrop of gentle slopes. Use caution off the tee to avoid the lake on the left before tackling the uphill fairway."
create_hole 15 "Hole 15" "Par 3" "115" "An intimate hole tucked into a tight pocket of the landscape. The narrow green is guarded by bunkers and blends seamlessly with the surrounding natural contours."
create_hole 16 "Hole 16" "Par 4" "301" "Risk-reward par 4 with an initial corridor that opens into a wider landing area; strategy is key on this scenic fairway with a lake to the left. Avoid the two protective bunkers guarding the right side of the green for a successful finish."
create_hole 17 "Hole 17" "Par 5" "481" "The longest challenge on the course.  Positioning on the fairway is critical, navigate rolling fairways and open wind exposure as you approach a green bordered by raw, natural hazards."
create_hole 18 "Hole 18" "Par 4" "307" "A grand finishing hole with the clubhouse as your backdrop. Watch for the small creek crossing the fairway 100m before the green and account for the wind on your final putts."

echo "Finished seeding 18 holes."
