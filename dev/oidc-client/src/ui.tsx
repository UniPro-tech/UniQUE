/** @jsxImportSource hono/jsx */

import { raw } from "hono/html";
import { clientId, issuer } from "./config";
import type { LoginState } from "./types";

const clientScript = `
  const setStage = (stage) => {
    document.querySelectorAll('[data-step]').forEach((node) =>
      node.classList.toggle('active', Number(node.dataset.step) <= stage),
    );
    document.querySelectorAll('[data-edge]').forEach((node) =>
      node.classList.toggle('active', Number(node.dataset.edge) < stage),
    );
  };
  document.getElementById('start-login')?.addEventListener('click', () => {
    const prompt = document.getElementById('prompt')?.value;
    const loginUrl = prompt ? '/login?prompt=' + encodeURIComponent(prompt) : '/login';
    const popup = window.open(loginUrl, 'unique-oidc-provider', 'popup,width=900,height=760');
    if (!popup) return window.location.assign(loginUrl);
    setStage(2);
  });
  window.addEventListener('message', (event) => {
    if (event.origin !== window.location.origin || event.data?.type !== 'unique-oidc-stage') return;
    setStage(event.data.stage);
    if (event.data.stage === 4) setTimeout(() => window.location.reload(), 700);
  });
`;

const commonSteps = [
  "PKCE request",
  "Sign in & consent",
  "Code exchange",
  "Verify tokens",
];

function Step({ index, label, active }: { index: number; label: string; active: boolean }) {
  const x = 60 + index * 125;
  return (
    <g data-step={index + 1} class={active ? "active" : ""}>
      <circle cx={x} cy="70" r="27" />
      <text x={x} y="75">{index + 1}</text>
      <foreignObject x={5 + index * 125} y="104" width="110" height="24">
        <div>{label}</div>
      </foreignObject>
    </g>
  );
}

function Branch({ active, cx, cy, label, step }: { active: boolean; cx: number; cy: number; label: string; step: string }) {
  return (
    <g class={active ? "active" : ""}>
      <circle cx={cx} cy={cy} r="27" />
      <text x={cx} y={cy + 5}>{step}</text>
      <foreignObject x={cx - 60} y={cy + 34} width="120" height="24">
        <div>{label}</div>
      </foreignObject>
    </g>
  );
}

function FlowChart({ session }: { session?: LoginState }) {
  const stage = session?.stage ?? 0;
  const rotated = session?.path === "rotated";
  const directRevoked = stage === 6 && session?.path === "direct";
  const rotatedRevoked = stage === 6 && rotated;

  return (
    <section class="flow">
      <svg viewBox="0 0 800 220" role="img" aria-label="OIDC flow">
        <text class="branch-label" x="575" y="20">Choose one path after token verification</text>
        {commonSteps.slice(0, -1).map((_, index) => (
          <path
            data-edge={index + 1}
            class={stage > index + 1 ? "active" : ""}
            d={`M${85 + index * 125} 70H${160 + index * 125}`}
          />
        ))}
        {commonSteps.map((label, index) => (
          <Step index={index} label={label} active={stage >= index + 1} />
        ))}
        <path class={rotated ? "active" : ""} d="M455 58L533 45" />
        <path class={directRevoked ? "active" : ""} d="M455 82L533 145" />
        <path class={rotatedRevoked ? "active" : ""} d="M587 45H668" />
        <Branch active={rotated} cx={560} cy={45} label="Refresh rotation" step="5A" />
        <Branch active={directRevoked} cx={560} cy={145} label="Direct revocation" step="5B" />
        <Branch active={rotatedRevoked} cx={695} cy={45} label="Revoke rotated set" step="6" />
      </svg>
    </section>
  );
}

function PromptControls({ session }: { session?: LoginState }) {
  return (
    <>
      <label class="prompt">
        prompt
        <select id="prompt">
          <option value="" selected={!session?.prompt}>default</option>
          <option value="consent" selected={session?.prompt === "consent"}>consent</option>
          <option value="none" selected={session?.prompt === "none"}>none</option>
        </select>
      </label>
      <button id="start-login">Start OIDC login</button>
      {session && (
        <form method="post" action="/reset">
          <button class="secondary">Reset test session</button>
        </form>
      )}
    </>
  );
}

function Milestones({ session }: { session?: LoginState }) {
  if (!session?.events.length) return null;
  return (
    <section class="milestones">
      <h2>Flow log</h2>
      {session.events.map((event, index) => (
        <details open={index === session.events.length - 1}>
          <summary><span>{index + 1}</span>{event.title}</summary>
          <pre>{JSON.stringify(event.details, null, 2)}</pre>
        </details>
      ))}
    </section>
  );
}

function TokenActions({ session }: { session?: LoginState }) {
  if (!session?.tokens || session.error || session.stage >= 6) return null;
  return (
    <>
      <form method="post" action="/refresh"><button>Rotate Refresh Token</button></form>
      <form method="post" action="/revoke"><button class="danger">Revoke Current Token Set</button></form>
    </>
  );
}

export function Page({ title, session, children, callbackComplete = false, callbackError = false }: { title: string; session?: LoginState; children: unknown; callbackComplete?: boolean; callbackError?: boolean }) {
  const callbackEvents = JSON.stringify(session?.events ?? []).replace(/</g, "\\u003c");
  const callbackScript = callbackComplete
    ? `<script>if(window.opener){window.opener.postMessage({type:"unique-oidc-stage",stage:3,outcome:"success",events:${callbackEvents}},window.location.origin);setTimeout(()=>window.opener.postMessage({type:"unique-oidc-stage",stage:4,outcome:"success",events:${callbackEvents}},window.location.origin),500);setTimeout(()=>window.close(),900)}</script>`
    : callbackError
      ? `<script>if(window.opener){window.opener.postMessage({type:"unique-oidc-stage",stage:4,outcome:"error",events:${callbackEvents}},window.location.origin);setTimeout(()=>window.close(),400)}</script>`
    : "";
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <title>{title}</title>
        <link rel="stylesheet" href="/styles.css" />
      </head>
      <body>
        <FlowChart session={session} />
        <h1>{title}</h1>
        <PromptControls session={session} />
        {children}
        <Milestones session={session} />
        <TokenActions session={session} />
        <script>{raw(clientScript)}</script>
        {raw(callbackScript)}
      </body>
    </html>
  );
}

export function Home({ session }: { session?: LoginState }) {
  return (
    <Page title="UniQUE OIDC Test Client" session={session}>
      {session ? (
        <p>Current flow state is retained in this browser session.</p>
      ) : (
        <>
          <p>Issuer: {issuer}</p>
          <p>Client ID: {clientId}</p>
          <noscript><p><a href="/login">Start OIDC login without flow visualization</a></p></noscript>
        </>
      )}
    </Page>
  );
}

export const ErrorPage = ({ message, session, callbackError = false }: { message: string; session?: LoginState; callbackError?: boolean }) => (
  <Page title="OIDC test failed" session={session} callbackError={callbackError}><pre>{message}</pre></Page>
);

export const shortToken = (token: string) =>
  `${token.slice(0, 16)}...${token.slice(-12)}`;
