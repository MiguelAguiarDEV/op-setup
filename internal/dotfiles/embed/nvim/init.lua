-- Neovim config
-- =====================================

-- Leader key
vim.g.mapleader = " "
vim.g.maplocalleader = " "

-- Opciones generales
local opt = vim.opt

opt.number = true           -- Números de línea
opt.relativenumber = true   -- Números relativos
opt.cursorline = true       -- Resaltar línea actual
opt.signcolumn = "yes"      -- Columna de signos siempre visible

opt.tabstop = 4             -- Tab = 4 espacios
opt.shiftwidth = 4          -- Indentación = 4 espacios
opt.expandtab = true        -- Usar espacios en vez de tabs
opt.smartindent = true      -- Indentación inteligente

opt.wrap = false            -- No hacer wrap de líneas largas
opt.scrolloff = 8           -- Mantener 8 líneas de contexto
opt.sidescrolloff = 8

opt.ignorecase = true       -- Búsqueda sin importar mayúsculas
opt.smartcase = true        -- ...a menos que uses mayúsculas
opt.hlsearch = true         -- Resaltar resultados de búsqueda
opt.incsearch = true        -- Búsqueda incremental

opt.splitbelow = true       -- Splits nuevos abajo
opt.splitright = true       -- Splits nuevos a la derecha

opt.termguicolors = true    -- Colores 24-bit
opt.background = "dark"

opt.undofile = true         -- Undo persistente
opt.swapfile = false        -- Sin archivos swap
opt.backup = false          -- Sin backups

opt.updatetime = 250        -- Más rápido para CursorHold
opt.timeoutlen = 300        -- Timeout para secuencias de teclas

opt.clipboard = "unnamedplus" -- Usar clipboard del sistema
opt.mouse = "a"             -- Mouse habilitado

opt.showmode = false        -- No mostrar modo (ya se ve en statusline)
opt.laststatus = 3          -- Statusline global

opt.autoread = true         -- Recargar archivos modificados externamente (opencode.nvim)

-- Keymaps útiles
local map = vim.keymap.set

-- Limpiar búsqueda con Esc
map("n", "<Esc>", "<cmd>nohlsearch<CR>")

-- Moverse entre ventanas con Ctrl+hjkl
map("n", "<C-h>", "<C-w>h")
map("n", "<C-j>", "<C-w>j")
map("n", "<C-k>", "<C-w>k")
map("n", "<C-l>", "<C-w>l")

-- Mover líneas en modo visual
map("v", "J", ":m '>+1<CR>gv=gv")
map("v", "K", ":m '<-2<CR>gv=gv")

-- Mantener cursor centrado al scrollear
map("n", "<C-d>", "<C-d>zz")
map("n", "<C-u>", "<C-u>zz")
map("n", "n", "nzzzv")
map("n", "N", "Nzzzv")

-- Mejor pegar sobre selección (no pierde el registro)
map("x", "<leader>p", '"_dP')

-- Explorador de archivos nativo
map("n", "<leader>e", "<cmd>Explore<CR>")

-- Guardar y salir rápidos
map("n", "<leader>w", "<cmd>w<CR>")
map("n", "<leader>q", "<cmd>q<CR>")

-- Buffers
map("n", "<S-h>", "<cmd>bprevious<CR>")
map("n", "<S-l>", "<cmd>bnext<CR>")

-- Diagnósticos
map("n", "[d", vim.diagnostic.goto_prev)
map("n", "]d", vim.diagnostic.goto_next)

-- opencode.nvim keymaps
map({ "n", "x" }, "<leader>oa", function() require("opencode").ask("@this: ", { submit = true }) end, { desc = "Ask opencode" })
map({ "n", "x" }, "<leader>os", function() require("opencode").select() end, { desc = "Select opencode action" })
map({ "n", "t" }, "<leader>oo", function() require("opencode").toggle() end, { desc = "Toggle opencode" })
map({ "n", "x" }, "<leader>or", function() return require("opencode").operator("@this ") end, { desc = "Add range to opencode", expr = true })

-- =====================================
-- Bootstrap lazy.nvim
-- =====================================
local lazypath = vim.fn.stdpath("data") .. "/lazy/lazy.nvim"
if not (vim.uv or vim.loop).fs_stat(lazypath) then
    local lazyrepo = "https://github.com/folke/lazy.nvim.git"
    local out = vim.fn.system({
        "git", "clone", "--filter=blob:none", "--branch=stable", lazyrepo, lazypath,
    })
    if vim.v.shell_error ~= 0 then
        vim.api.nvim_echo({
            { "Failed to clone lazy.nvim:\n", "ErrorMsg" },
            { out, "WarningMsg" },
            { "\nPress any key to exit..." },
        }, true, {})
        vim.fn.getchar()
        os.exit(1)
    end
end
vim.opt.rtp:prepend(lazypath)

-- Load plugins from lua/plugins/
require("lazy").setup("plugins", {
    defaults = { lazy = false },
    checker = { enabled = false },
    change_detection = { notify = false },
})
