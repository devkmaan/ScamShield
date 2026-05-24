export async function api(path, options = {}) {
  const response = await fetch(path, {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {})
    },
    ...options
  });
  const text = await response.text();
  let payload = null;
  try {
    payload = text ? JSON.parse(text) : {};
  } catch {
    payload = text;
  }
  if (!response.ok) {
    const message = typeof payload === 'string' ? payload : payload?.error || 'Request failed';
    throw new Error(message);
  }
  return payload;
}

export function jsonBody(value) {
  return {
    method: 'POST',
    body: JSON.stringify(value)
  };
}

