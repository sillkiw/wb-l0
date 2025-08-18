(() => {
  const $ = (id) => document.getElementById(id);

  const form = $('searchForm');
  const input = $('orderInput');
  const btn = $('searchBtn');
  const statusEl = $('status');
  const metaEl = $('meta');
  const resultEl = $('result');
  const jsonEl = $('json');
  const summaryEl = $('summary');
  const historyWrap = $('history');
  const historyList = $('historyList');

  const HISTORY_KEY = 'order_search_history';

  const setStatus = (msg) => statusEl.textContent = msg || '';
  const setMeta = (msg) => metaEl.textContent = msg || '';

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));
  }

  function keepHistory(id) {
    try {
      const arr = JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]');
      if (!id) return;
      const next = [id, ...arr.filter(x => x !== id)].slice(0, 6);
      localStorage.setItem(HISTORY_KEY, JSON.stringify(next));
      renderHistory(next);
    } catch {}
  }

  function renderHistory(arr) {
    if (!arr || arr.length === 0) { historyWrap.hidden = true; return; }
    historyWrap.hidden = false;
    historyList.innerHTML = '';
    arr.forEach(id => {
      const span = document.createElement('span');
      span.className = 'pill';
      span.textContent = id;
      span.title = 'Нажмите, чтобы подставить в поиск';
      span.onclick = () => { input.value = id; form.requestSubmit(); };
      historyList.appendChild(span);
    });
  }

  // первичная история
  renderHistory(JSON.parse(localStorage.getItem(HISTORY_KEY) || '[]'));

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = (input.value || '').trim();
    if (!id) { setStatus('Введите корректный order_id.'); return; }

    btn.disabled = true;
    setStatus('Идёт запрос…');
    setMeta('');
    resultEl.hidden = true;
    summaryEl.hidden = true;
    jsonEl.textContent = '';

    const t0 = performance.now();
    const ac = new AbortController();
    const timeout = setTimeout(() => ac.abort(), 5000); // 5s таймаут

    try {
      const resp = await fetch(`/order/${encodeURIComponent(id)}`, {
        headers: { 'Accept': 'application/json' },
        signal: ac.signal,
      });
      clearTimeout(timeout);

      const src = resp.headers.get('X-Source') || 'неизвестно';
      const dt = (performance.now() - t0).toFixed(0) + ' ms';

      if (resp.status === 404) {
        setStatus('Заказ не найден.');
        setMeta(`Источник: ${src} · Время: ${dt}`);
        return;
      }
      if (!resp.ok) {
        const text = await resp.text().catch(() => '');
        setStatus(`Ошибка сервера (${resp.status}). ${text || ''}`);
        setMeta(`Источник: ${src} · Время: ${dt}`);
        return;
      }

      const data = await resp.json();
      setStatus('Готово.');
      setMeta(`Источник: ${src} · Время: ${dt}`);
      resultEl.hidden = false;

      // краткое резюме
      try {
        const uid = data.order_uid || data.id || '';
        const items = Array.isArray(data.items) ? data.items.length :
                      (Array.isArray(data.order_items) ? data.order_items.length : undefined);
        const buyer = (data.delivery && data.delivery.name) || (data.customer && data.customer.name) || '';
        const amount = data.payment?.amount ?? data.total_amount ?? data.amount;

        const cells = [];
        if (uid)    cells.push(`<div class="k">Order ID</div><div>${escapeHtml(uid)}</div>`);
        if (buyer)  cells.push(`<div class="k">Получатель</div><div>${escapeHtml(buyer)}</div>`);
        if (items!=null) cells.push(`<div class="k">Товаров</div><div>${items}</div>`);
        if (amount!=null) cells.push(`<div class="k">Сумма</div><div>${amount}</div>`);

        if (cells.length) {
          summaryEl.innerHTML = cells.join('');
          summaryEl.hidden = false;
        }
      } catch {}

      jsonEl.textContent = JSON.stringify(data, null, 2);
      keepHistory(id);
    } catch (err) {
      const aborted = err && (err.name === 'AbortError');
      setStatus(aborted ? 'Таймаут запроса (5s).' : `Сетевая ошибка: ${err}`);
    } finally {
      clearTimeout(timeout);
      btn.disabled = false;
    }
  });

  // Enter в поле = submit
  input.addEventListener('keydown', (e) => {
    if (e.key === 'Enter') form.requestSubmit();
  });
})();
