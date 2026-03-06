# Установка базового образа
FROM node:18-alpine

# Установка рабочей директории
WORKDIR /app

# Копирование package.json и package-lock.json
COPY package.json package-lock.json ./

# Установка зависимостей
RUN npm install

# Копирование остальных файлов проекта
COPY . .

# Сборка приложения
RUN npm run build

# Установка статического сервера
RUN npm install -g serve

# Запуск приложения
CMD ["serve", "-s", "build", "-l", "3000"]