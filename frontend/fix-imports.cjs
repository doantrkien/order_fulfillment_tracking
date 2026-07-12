const fs = require('fs');
const path = require('path');

function replaceInFile(filePath, replacements) {
    let content = fs.readFileSync(filePath, 'utf8');
    for (const [pattern, replacement] of replacements) {
        content = content.replace(pattern, replacement);
    }
    fs.writeFileSync(filePath, content);
    console.log(`Updated ${filePath}`);
}

// 1. Remove Placeholder from App.tsx
replaceInFile('src/App.tsx', [
    [/const Placeholder = \(\{\s*title\s*\}\: \{\s*title\: string\s*\}\) => \([\s\S]*?\);\n\n/g, '']
]);

// 2. Add 'packed' to OrderStatus in order.service.ts
replaceInFile('src/services/order.service.ts', [
    [/'created' \| 'paid' \| 'shipped'/, "'created' | 'paid' | 'packed' | 'shipped'"]
]);

// 3. Fix imports
const typeFixes = [
    ['src/components/ai/DraftPanel.tsx', /import \{ aiService, AIDraftResponse \}/, 'import { aiService, type AIDraftResponse }'],
    ['src/components/ai/ExceptionPanel.tsx', /import \{ AIExceptionAnalysis \}/, 'import type { AIExceptionAnalysis }'],
    ['src/components/driver/DriverOrderCard.tsx', /import \{ Order \}/, 'import type { Order }'],
    ['src/components/orders/OrderFilters.tsx', /import \{ OrderQuery, OrderStatus \}/, 'import type { OrderQuery, OrderStatus }'],
    ['src/components/orders/OrderStatusModal.tsx', /import \{ OrderStatus \}/, 'import type { OrderStatus }'],
    ['src/components/orders/OrderTable.tsx', /import \{ Order \}/, 'import type { Order }'],
    ['src/components/orders/StatusBadge.tsx', /import \{ OrderStatus \}/, 'import type { OrderStatus }'],
    ['src/contexts/AuthContext.tsx', /import React, \{ createContext, useContext, useState, useEffect, ReactNode \}/, 'import React, { createContext, useContext, useState, useEffect, type ReactNode }'],
    ['src/contexts/AuthContext.tsx', /import \{ login, logout, LoginCredentials, User \}/, 'import { login, logout, type LoginCredentials, type User }'],
    ['src/hooks/useOrders.ts', /import \{ orderService, Order, OrderQuery, PaginatedOrders \}/, 'import { orderService, type OrderQuery, type PaginatedOrders }'], // removed Order entirely since it's unused
    ['src/pages/admin/DailyReportPage.tsx', /import \{ reportService, DailyReportResponse \}/, 'import { reportService, type DailyReportResponse }'],
    ['src/pages/admin/OrderDetailPage.tsx', /import \{ orderService, Order \}/, 'import { orderService, type Order }'],
    ['src/pages/admin/OrderDetailPage.tsx', /import \{ aiService, AIExceptionAnalysis \}/, 'import { aiService, type AIExceptionAnalysis }'],
    ['src/pages/driver/DriverOrderDetailPage.tsx', /import \{ orderService, Order \}/, 'import { orderService, type Order }'],
    ['src/services/event.service.ts', /import \{ OrderStatus \}/, 'import type { OrderStatus }'],
    ['src/services/report.service.ts', /import \{ ApiResponse \}/, 'import type { ApiResponse }'],
];

for (const [file, pattern, replacement] of typeFixes) {
    if (fs.existsSync(file)) {
        replaceInFile(file, [[pattern, replacement]]);
    } else {
        console.warn(`File not found: ${file}`);
    }
}

console.log('Fixes applied.');
