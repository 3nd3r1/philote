// Posts the generated forms the way a browser without JavaScript would:
// the form's hidden action fields plus the typed values, as multipart.
// Usage: node smoke-forms.mjs <web url> <a horse id>
const [web, horse] = process.argv.slice(2);
let fails = 0;

function decode(s) {
    return s
        .replaceAll("&quot;", '"')
        .replaceAll("&#x27;", "'")
        .replaceAll("&lt;", "<")
        .replaceAll("&gt;", ">")
        .replaceAll("&amp;", "&");
}

async function post(path, fields) {
    const html = await (await fetch(web + path)).text();
    const data = new FormData();
    for (const [, tag] of html.matchAll(/(<input[^>]*name="\$ACTION[^>]*>)/g)) {
        const name = decode(tag.match(/name="([^"]*)"/)[1]);
        const value = tag.match(/value="([^"]*)"/);
        data.append(name, value ? decode(value[1]) : "");
    }
    for (const [name, value] of Object.entries(fields)) {
        data.append(name, value);
    }
    return fetch(web + path, { method: "POST", body: data, redirect: "manual" });
}

function check(ok, what) {
    console.log(`${ok ? "ok  " : "FAIL"} form ${what}`);
    if (!ok) fails++;
}

let res = await post("/horses/new", { name: "Black Beauty", breed: "arabian", age: "6", bio: "Gentle." });
const created = res.headers.get("location") ?? "";
check(res.status === 303 && /^\/horses\/[0-9a-f-]{36}$/.test(created), `/horses/new saves and goes to the horse (${res.status} ${created})`);

res = await post("/horses/new", { name: "Old Timer", breed: "other", age: "99", bio: "Too old." });
check(res.status === 200 && (await res.text()).includes("must be at most 40"), "/horses/new shows the backend's issue under age");

res = await post(`/horses/${horse}/edit`, { name: "Sea Biscuit", breed: "mustang", age: "8", bio: "Edited by a form." });
check(res.status === 303 && res.headers.get("location") === `/horses/${horse}`, `/horses/[id]/edit saves (${res.status})`);
const page = await (await fetch(`${web}/horses/${horse}`)).text();
check(page.includes("Edited by a form."), "/horses/[id] shows the edit");

res = await post("/swipe", { horse, target: horse, liked: "on" });
check(res.status === 200 && (await res.text()).includes("Same horse."), "/swipe shows a step's declared error");

if (fails > 0) {
    console.log(`forms: ${fails} failed`);
    process.exit(1);
}
