import './style.css';
import { Analyze, ReadFile, SelectFile } from '../wailsjs/go/main/App';

const app = document.getElementById('app');

app.innerHTML = `
  <h1>Метрика граничных значений (Ruby)</h1>

  <div class="layout">
    <div class="panel">
      <h2>Исходный код Ruby</h2>
      <textarea id="source" placeholder="Вставь Ruby-код или открой файл .rb"></textarea>
      <div class="buttons">
        <button id="btn-analyze">Анализировать</button>
        <button id="btn-open">Открыть .rb</button>
        <button id="btn-print">Печать отчёта</button>
      </div>
    </div>
    <div class="panel">
      <h2>Результат анализа</h2>
      <div id="result" class="result">
        <p class="muted">Введите код и нажмите «Анализировать».</p>
      </div>
    </div>
  </div>
`;

const $source = document.getElementById('source');
const $result = document.getElementById('result');

document.getElementById('btn-analyze').addEventListener('click', runAnalysis);

document.getElementById('btn-open').addEventListener('click', async () => {
    try {
        const path = await SelectFile();
        if (!path) return;
        $source.value = await ReadFile(path);
        runAnalysis();
    } catch (e) {
        $result.innerHTML = `<p class="error">Ошибка открытия: ${escapeHtml(e)}</p>`;
    }
});

document.getElementById('btn-print').addEventListener('click', () => window.print());

async function runAnalysis() {
    const src = $source.value;
    if (!src.trim()) {
        $result.innerHTML = '<p class="muted">Введите Ruby-код.</p>';
        return;
    }

    try {
        const res = await Analyze(src);
        if (!res || !res.boundary) {
            $result.innerHTML = `<p class="error">Бэкенд не вернул res.boundary. Проверьте Go-код.</p>`;
            return;
        }
        renderBoundary(res.boundary);
    } catch (e) {
        $result.innerHTML = `<p class="error">Ошибка: ${escapeHtml(e)}</p>`;
    }
}

function renderBoundary(b) {
    const subs = b.subs || [];
    let vertices = b.vertices || [];
    const nu = vertices.length;

    // Сортируем вершины по ID на всякий случай, чтобы они шли строго по порядку
    vertices.sort((a, b) => a.vertexId - b.vertexId);

    // Таблица 1 (остается без изменений)
    const subsRows = subs.map(s => `
      <tr>
        <td>${s.selectVertex}</td>
        <td>${(s.transitions || []).join(', ')}</td>
        <td class="num">${s.selectCost}</td>
        <td>${(s.subgraphVerts || []).join(', ')}</td>
        <td>${s.lowerBound}</td>
      </tr>
    `).join('');

    // Таблица 2 (Делаем в две колонки)
    const mid = Math.ceil(nu / 2);
    let vertRows = '';

    for (let i = 0; i < mid; i++) {
        const v1 = vertices[i];
        const v2 = vertices[i + mid];

        vertRows += `
          <tr>
            <td>${v1.vertexId}</td>
            <td class="num">${v1.complexity}</td>
            <td>${v2 ? v2.vertexId : ''}</td>
            <td class="num">${v2 ? v2.complexity : ''}</td>
          </tr>
        `;
    }

    $result.innerHTML = `
      <div class="cards">
        <div class="metric-card">
          <span>Sa — абсолютная сложность</span>
          <b>${b.sa}</b>
        </div>
        <div class="metric-card">
          <span>So — относительная сложность</span>
          <b>${Number(b.so).toFixed(3)}</b>
        </div>
        <div class="metric-card">
          <span>Вершин ν</span>
          <b>${nu}</b>
        </div>
        <div class="metric-card">
          <span>Выбирающих вершин</span>
          <b>${subs.length}</b>
        </div>
      </div>

      <div class="formula">
        <b>So</b> = 1 − (ν − 1) / Sa
        <br>
        So = 1 − (${nu} − 1) / ${b.sa} = ${Number(b.so).toFixed(3)}
      </div>

      <h3>Таблица 1. Свойства подграфов программы</h3>
      <table>
        <thead>
          <tr>
            <th>№ вершины выбора</th>
            <th>№№ вершин перехода</th>
            <th>Скорр. сложность выбора</th>
            <th>№№ вершин подграфа</th>
            <th>№ нижней границы</th>
          </tr>
        </thead>
        <tbody>
          ${subsRows || '<tr><td colspan="5" class="muted">Нет выбирающих вершин</td></tr>'}
        </tbody>
      </table>

      <h3>Исправленная таблица 2. Скорректированные сложности вершин</h3>
      <p class="muted">Для графа из ${nu} вершин получаем следующие значения:</p>
      <table>
        <thead>
          <tr>
            <th>№ вершины</th>
            <th>Скорр. сложность</th>
            <th>№ вершины</th>
            <th>Скорр. сложность</th>
          </tr>
        </thead>
        <tbody>
          ${vertRows || '<tr><td colspan="4" class="muted">—</td></tr>'}
        </tbody>
      </table>
    `;
}

function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, c => ({
        '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
    })[c]);
}