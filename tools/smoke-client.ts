// Calls the running backend through the generated Next.js client. smoke.sh
// copies it into the built web part and runs it with API_URL set.
import { horses, matches, swipe } from "@/lib/api";

let fails = 0;
function expect(what: string, ok: boolean, got: unknown) {
    console.log(`${ok ? "ok  " : "FAIL"} client ${what}${ok ? "" : ": " + JSON.stringify(got)}`);
    if (!ok) fails++;
}

const shadow = await horses.create({ name: "Shadowfax", breed: "other", age: 30, bio: "Fast." });
expect("creates a horse", shadow.ok && shadow.value.handle === "shadowfax", shadow);
const bad = await horses.create({ name: "Nobody", breed: "other", age: 0, bio: "" });
expect("returns invalid_input with issues", !bad.ok && bad.error === "invalid_input" && (bad.issues?.length ?? 0) > 0, bad);
const missing = await horses.get("nope");
expect("returns not_found", !missing.ok && missing.error === "not_found", missing);

const pony = await horses.create({ name: "Pony", breed: "shetland", age: 4, bio: "Small." });
if (!shadow.ok || !pony.ok) throw new Error("setup failed");
const self = await swipe({ horse: pony.value.id, target: pony.value.id, liked: true });
expect("returns a step's declared error", !self.ok && self.error === "same_horse", self);
await swipe({ horse: pony.value.id, target: shadow.value.id, liked: true });
const back = await swipe({ horse: shadow.value.id, target: pony.value.id, liked: true });
expect("swipes", back.ok && back.value.liked, back);
const liked = await horses.swipedBy(pony.value.id);
expect("lists a has-many relation", liked.ok && liked.value.length === 1 && liked.value[0].horse_id === shadow.value.id, liked);
const all = await matches.list();
const mine = all.ok ? all.value.filter((m) => [m.first_id, m.second_id].includes(pony.value.id)) : [];
expect("sees the match", mine.length === 1, all);
const loner = await horses.create({ name: "Loner", breed: "arabian", age: 9, bio: "Keeps to itself." });
if (!loner.ok) throw new Error("setup failed");
const gone = await horses.delete(loner.value.id);
expect("deletes", gone.ok, gone);

if (fails > 0) {
    console.log(`smoke: ${fails} client calls failed`);
    process.exit(1);
}
