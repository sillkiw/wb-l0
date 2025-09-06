(() => {
  const $ = (id) => document.getElementById(id);

  const form = $('searchForm');
  const input = $('orderInput');
  const btn = $('searchBtn');
  const statusEl = $('status');
  const metaEl = $('meta');
  const resultEl = $('result');
  const summaryEl = $('summary');
  const itemsEl = $('items');
  const historyWrap = $('history');
  const historyList = $('historyList');

  // история только в текущей сессии
  const HISTORY_KEY = 'order_search_history_session';

  const setStatus = (msg) => statusEl.textContent = msg || '';
  const setMeta = (msg) => metaEl.textContent = msg || '';

  const esc = (s) =>
    String(s ?? '').replace(/[&<>"']/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));

  const keepHistory = (id) => {
    try {
      const arr = JSON.parse(sessionStorage.getItem(HISTORY_KEY) || '[]');
      if (!id) return;
      const next = [id, ...arr.filter(x => x !== id)].slice(0, 6);
      sessionStorage.setItem(HISTORY_KEY, JSON.stringify(next));
      renderHistory(next);
    } catch {}
  };

  const renderHistory = (arr) => {
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
  };

  // начальная история
  renderHistory(JSON.parse(sessionStorage.getItem(HISTORY_KEY) || '[]'));

  // ---------- форматтеры ----------
  const maskPhone = (p) => (p || '').replace(/(\+?\d{1,3})?(\d{3})(\d{2})(\d{2})(\d{2})$/, (_, c, a, b, d, e) =>
    `${c || ''} ${a}-**-**-${e}`);

  const formatAmt = (n, ccy) => {
    if (n == null) return '';
    try {
      const map = { RUB: 'ru-RU', KZT: 'kk-KZ', BYN: 'be-BY', KGS: 'ru-KG', AMD: 'hy-AM', TRY: 'tr-TR', UZS: 'uz-UZ', AZN: 'az-AZ', GEL: 'ka-GE' };
      const loc = map[ccy] || 'ru-RU';
      return new Intl.NumberFormat(loc, { style: 'currency', currency: ccy || 'RUB', maximumFractionDigits: 0 }).format(n);
    } catch { return `${n} ${ccy || ''}`; }
  };

  const fmtDateIso = (s) => {
    if (!s) return '';
    const d = new Date(s);
    if (isNaN(d)) return esc(s);
    return d.toLocaleString(undefined, { year:'numeric', month:'short', day:'2-digit', hour:'2-digit', minute:'2-digit' });
  };

  const fmtUnix = (secs) => {
    if (secs == null) return '';
    const d = new Date(Number(secs) * 1000);
    if (isNaN(d)) return String(secs);
    return d.toLocaleString(undefined, { year:'numeric', month:'short', day:'2-digit', hour:'2-digit', minute:'2-digit' });
  };

  // ---------- отрисовка ----------
  const renderSummary = (data) => {
    const uid = data.order_uid || data.id || '';
    const track = data.track_number || '';
    const entry = data.entry || '';
    const created = data.date_created || data.created_at || '';
    const delivery = data.delivery || {};
    const payment = data.payment || {};
    const items = Array.isArray(data.items) ? data.items : (Array.isArray(data.order_items) ? data.order_items : []);
    const itemsCount = items.length;

    const add = (k, v) => { if (v !== '' && v != null) cells.push(`<div class="k">${k}</div><div>${v}</div>`); };

    const cells = [];
    if (uid)     add('Order ID', `<code class="mono">${esc(uid)}</code>`);
    if (track)   add('Трек', esc(track));
    if (created) add('Создан', fmtDateIso(created));
    if (data.delivery_service) add('Служба', esc(data.delivery_service));

    if (delivery.name)    add('Получатель', esc(delivery.name));
    if (delivery.city)    add('Город', esc(delivery.city));
    if (delivery.address) add('Адрес', esc(delivery.address));
    if (delivery.phone)   add('Телефон', esc(maskPhone(delivery.phone)));
    if (entry)            add('Канал', esc(entry));

    // --- Оплата (разбивка) ---
    const ccy = payment.currency || 'RUB';
    if (payment.provider || payment.bank) {
      add('Оплата', esc([payment.provider, payment.bank].filter(Boolean).join(' · ')));
    }
    if (payment.goods_total != null)   add('Товары', esc(formatAmt(payment.goods_total, ccy)));
    if (payment.delivery_cost != null) add('Доставка', esc(formatAmt(payment.delivery_cost, ccy)));
    if (payment.custom_fee)            add('Пошлина', esc(formatAmt(payment.custom_fee, ccy)));
    if (payment.amount != null)        add('Итого', `<b>${esc(formatAmt(payment.amount, ccy))}</b>`);

    if (payment.transaction) add('Транзакция', `<code class="mono">${esc(payment.transaction)}</code>`);
    if (payment.request_id)  add('Request ID', `<code class="mono">${esc(payment.request_id)}</code>`);
    if (payment.payment_dt != null) add('Оплачено', esc(fmtUnix(payment.payment_dt)));

    summaryEl.innerHTML = cells.join('');
    summaryEl.hidden = cells.length === 0;
  };

  const renderItems = (data) => {
    const items = Array.isArray(data.items) ? data.items : (Array.isArray(data.order_items) ? data.order_items : []);
    if (!items.length) { itemsEl.hidden = true; itemsEl.innerHTML = ''; return; }

    const ccy = data.payment?.currency || '';
    const html = items.map((it, idx) => {
      const line1 = esc(it.name || `Товар ${idx+1}`);
      const line2 = [
        it.brand ? esc(it.brand) : '',
        it.size ? `разм. ${esc(it.size)}` : '',
        (it.sale != null && it.sale !== 0) ? `скидка ${esc(it.sale)}%` : '',
      ].filter(Boolean).join(' · ');
      const price = (it.total_price ?? it.price);
      return `
        <div class="item">
          <div class="item__left">
            <div class="item__name">${line1}</div>
            ${line2 ? `<div class="item__meta">${line2}</div>` : ''}
          </div>
          <div class="item__right">${esc(formatAmt(price, ccy))}</div>
        </div>
      `;
    }).join('');

    itemsEl.innerHTML = `<h3>Товары</h3>${html}`;
    itemsEl.hidden = false;
  };

  // ---------- submit ----------
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const id = (input.value || '').trim();
    if (!id) { setStatus('Введите корректный order_id.'); return; }

    btn.disabled = true;
    setStatus('Идёт запрос…');
    setMeta('');
    resultEl.hidden = true;
    summaryEl.hidden = true;
    itemsEl.hidden = true;
    summaryEl.innerHTML = '';
    itemsEl.innerHTML = '';

    const t0 = performance.now();
    const ac = new AbortController();
    const timeout = setTimeout(() => ac.abort(), 7000); // 7s

    try {
      const resp = await fetch(`/order/${encodeURIComponent(id)}`, {
        headers: { 'Accept': 'application/json' },
        signal: ac.signal,
      });
      clearTimeout(timeout);

      const src = resp.headers.get('X-Source'); // может быть null
      const dt = `${(performance.now() - t0).toFixed(0)} ms`;

      if (resp.status === 404) {
        setStatus('Заказ не найден.');
        setMeta([src ? `Источник: ${src}` : null, `Время: ${dt}`].filter(Boolean).join(' · '));
        return;
      }
      if (!resp.ok) {
        const text = await resp.text().catch(() => '');
        setStatus(`Ошибка сервера (${resp.status}). ${text || ''}`);
        setMeta([src ? `Источник: ${src}` : null, `Время: ${dt}`].filter(Boolean).join(' · '));
        return;
      }

      const data = await resp.json();
      renderSummary(data);
      renderItems(data);

      resultEl.hidden = false;
      setStatus('Готово.');
      setMeta([src ? `Источник: ${src}` : null, `Время: ${dt}`].filter(Boolean).join(' · '));

      keepHistory(id);
    } catch (err) {
      const aborted = err && (err.name === 'AbortError');
      setStatus(aborted ? 'Таймаут запроса.' : `Сетевая ошибка: ${err}`);
    } finally {
      clearTimeout(timeout);
      btn.disabled = false;
    }
  });

  // Enter = submit
  $('orderInput').addEventListener('keydown', (e) => {
    if (e.key === 'Enter') form.requestSubmit();
  });
})();
