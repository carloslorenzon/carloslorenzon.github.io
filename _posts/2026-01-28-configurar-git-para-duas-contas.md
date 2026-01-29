---
layout: post
title:  "Como Configurar Múltiplas Contas Git/GitHub com Chaves SSH Diferentes"
categories: blog hard-skills git
author: Carlos Lorenzon
comments: true
---

Este guia mostra como configurar automaticamente diferentes credenciais Git (email e chave SSH) baseado no repositório que você está trabalhando.

## 📋 Cenário

Você trabalha com:
- **Conta corporativa** (ex: EmpresaX) com um email e chave SSH
- **Conta pessoal** com outro email e outra chave SSH

E quer que o Git use automaticamente as credenciais corretas sem precisar configurar manualmente em cada repositório.

## 🎯 Objetivo

- Repositórios de `empresax` → usar email corporativo + chave SSH corporativa
- Outros repositórios → usar email pessoal + chave SSH pessoal
- Tudo **automático**, sem configuração manual!

## 📝 Pré-requisitos

1. Ter duas chaves SSH diferentes criadas:

```bash
ls ~/.ssh/*.pub
```
   
Exemplo de resultado:
```
~/.ssh/id_ed25519.pub           (chave corporativa)
~/.ssh/id_ed25519_personal.pub  (chave pessoal)
```

2. Git versão 2.36.0 ou superior (para suporte a `hasconfig`)
```bash
git --version
```

## 🔧 Passo 1: Criar Chaves SSH (se ainda não tiver)

### Chave Corporativa
```bash
ssh-keygen -t ed25519 -C "seu.email@empresa.com" -f ~/.ssh/id_ed25519
```

### Chave Pessoal
```bash
ssh-keygen -t ed25519 -C "seu.email@gmail.com" -f ~/.ssh/id_ed25519_personal
```

**Dica:** Pressione ENTER quando pedir senha (sem senha é mais prático para desenvolvimento local).

## 🔧 Passo 2: Configurar SSH

Edite ou crie o arquivo `~/.ssh/config`:

```bash
nano ~/.ssh/config
```

Adicione a seguinte configuração:

```ssh
# Configuração para repositórios corporativos (EmpresaX)
Host github.com-empresax
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519
    IdentitiesOnly yes

# Configuração padrão para outros repositórios (pessoal)
Host github.com
    HostName github.com
    User git
    IdentityFile ~/.ssh/id_ed25519_personal
    IdentitiesOnly yes

# Fallback para qualquer outro host SSH
Host *
    AddKeysToAgent yes
    UseKeychain yes  # ⚠️ Específico do macOS - remova em Linux/Windows
```

### 📌 O que isso faz?

- `github.com-empresax` → usa a chave `id_ed25519` (corporativa)
- `github.com` → usa a chave `id_ed25519_personal` (pessoal)
- `IdentitiesOnly yes` → força usar apenas a chave especificada

## 🔧 Passo 3: Criar arquivo de configuração Git para conta corporativa

Crie o arquivo `~/.gitconfig-empresax`:

```bash
nano ~/.gitconfig-empresax
```

Com o seguinte conteúdo (ajuste com seus dados):

```ini
[user]
    email = seu.email@empresa.com
    name = Seu Nome
    signingkey = /Users/seu-usuario/.ssh/id_ed25519

[url "git@github.com-empresax:empresax/"]
    insteadOf = git@github.com:empresax/
    insteadOf = https://github.com/empresax/

```

### 📌 O que isso faz?

- Define email, nome e chave de assinatura para conta corporativa
- `insteadOf` → reescreve URLs para usar `github.com-empresax` automaticamente

## 🔧 Passo 4: Configurar arquivo Git principal

Edite o arquivo `~/.gitconfig`:

```bash
nano ~/.gitconfig
```

Configure assim (ajuste com seus dados):

```ini
# ⚠️ Configuração padrão (pessoal) DEVE vir ANTES dos includeIf
[user]
    email = seu.email@gmail.com
    name = Seu Nome
    signingkey = /Users/seu-usuario/.ssh/id_ed25519_personal

# Configuração condicional para repositórios corporativos
# includeIf vem POR ÚLTIMO para sobrescrever as configurações padrão
[includeIf "hasconfig:remote.*.url:**empresax/**"]
    path = ~/.gitconfig-empresax
```

### 📌 O que isso faz?

- Define configurações pessoais como **padrão**
- `includeIf` com `hasconfig` → detecta se o remote contém `empresax`
- Se detectar → inclui `~/.gitconfig-empresax` (sobrescreve com config corporativa)

### ⚠️ IMPORTANTE: Ordem importa!

A seção `[user]` padrão **deve vir antes** dos `[includeIf]` para que os includes possam sobrescrever os valores.

## ✅ Passo 5: Adicionar chaves SSH às contas do GitHub

### Conta Corporativa
1. Copie a chave pública corporativa:

```bash
cat ~/.ssh/id_ed25519.pub | pbcopy
```

2. Acesse GitHub (conta corporativa) → Settings → SSH and GPG keys → New SSH key
3. Cole a chave e salve

### Conta Pessoal

1. Copie a chave pública pessoal:
```bash
cat ~/.ssh/id_ed25519_personal.pub | pbcopy
```

2. Acesse GitHub (conta pessoal) → Settings → SSH and GPG keys → New SSH key
3. Cole a chave e salve

## 🔧 Passo 6: Adicionar chaves ao ssh-agent

Adicione as chaves ao ssh-agent para que sejam carregadas automaticamente:

```bash
# Adicionar chave corporativa
ssh-add ~/.ssh/id_ed25519

# Adicionar chave pessoal
ssh-add ~/.ssh/id_ed25519_personal
```

### Criar aliases para facilitar (opcional)

Adicione ao seu `~/.zshrc` ou `~/.bashrc`:

```bash
# Aliases para adicionar chaves SSH
alias sshempresax="ssh-add ~/.ssh/id_ed25519"
alias sshpessoal="ssh-add ~/.ssh/id_ed25519_personal"
```

Depois recarregue:
```bash
source ~/.zshrc
```

## 🧪 Testando a Configuração

### Teste 1: Repositório Corporativo

```bash
# Clone ou entre em um repositório corporativo
cd seu-repo-corporativo

# Verifique as configurações
git config user.email
# Deve mostrar: seu.email@empresa.com

git config user.signingkey
# Deve mostrar: /Users/seu-usuario/.ssh/id_ed25519
```

### Teste 2: Repositório Pessoal

```bash
# Clone ou entre em um repositório pessoal
cd seu-repo-pessoal

# Verifique as configurações
git config user.email
# Deve mostrar: seu.email@gmail.com

git config user.signingkey
# Deve mostrar: /Users/seu-usuario/.ssh/id_ed25519_personal
```

### Teste 3: Conexão SSH

Teste a conexão SSH com ambas as chaves:

```bash
# Teste chave corporativa
ssh -T git@github.com-empresax

# Teste chave pessoal
ssh -T git@github.com
```

Ambos devem retornar uma mensagem de sucesso do GitHub.

## 🚀 Como Usar no Dia a Dia

### Clonando Repositórios

**Repositórios corporativos:**

```bash
git clone git@github.com:empresax/seu-repo.git
```

**Repositórios pessoais:**

```bash
git clone git@github.com:seu-usuario/seu-repo.git
```

### Fazendo Commits

Não muda nada! O Git automaticamente usa:

- Email correto
- Chave SSH correta para assinatura
- Chave SSH correta para push/pull

```bash
git add .
git commit -m "Meu commit"
git push
```

## 🔍 Comandos Úteis para Debug

### Ver qual configuração está sendo usada

```bash
# Ver de onde vem cada configuração
git config --list --show-origin | grep user

# Ver configurações do repositório atual
git config user.email
git config user.signingkey
git config --list --show-scope | grep user
```

### Testar qual chave SSH está sendo usada

```bash
# Ver quais chaves estão carregadas
ssh-add -l

# Testar conexão com verbose
ssh -vT git@github.com
ssh -vT git@github.com-empresax
```

### Ver logs de commits com informações de assinatura

```bash
git log --show-signature -1
```

## 🎯 Personalizando para Outros Casos

### Adicionar mais organizações

No `~/.gitconfig`, adicione mais `includeIf`:

```ini
[includeIf "hasconfig:remote.*.url:**sua-empresa/**"]
    path = ~/.gitconfig-empresa
[includeIf "hasconfig:remote.*.url:**outro-org/**"]
    path = ~/.gitconfig-empresa
```

### Usar baseado em pasta ao invés de URL

Se você organiza repos por pasta, pode usar `gitdir`:

```ini
# Todos repos em ~/projetos/empresa/ usam config corporativa
[includeIf "gitdir:~/projetos/empresa/"]
    path = ~/.gitconfig-empresa

# Todos repos em ~/projetos/pessoal/ usam config pessoal
[includeIf "gitdir:~/projetos/pessoal/"]
    path = ~/.gitconfig-pessoal
```

## ❓ Troubleshooting

### Problema: Email errado nos commits

**Causa:** Repositório foi clonado antes da configuração

**Solução:** 
```bash
cd seu-repositorio
git config user.email seu.email@correto.com
git config user.signingkey /caminho/para/chave/correta
```

### Problema: Chave SSH não funciona

**Causa:** Chave não está carregada no ssh-agent

**Solução:**
```bash
# Adicionar chave ao ssh-agent
ssh-add ~/.ssh/id_ed25519
ssh-add ~/.ssh/id_ed25519_personal

# Verificar
ssh-add -l
```

### Problema: `includeIf` não funciona

**Causa:** Versão antiga do Git

**Solução:**
```bash
# Verificar versão (precisa 2.36+)
git --version

# Atualizar Git (macOS)
brew upgrade git
```

### Problema: "Permission denied (publickey)"

**Causa:** Chave não foi adicionada ao GitHub

**Solução:**
1. Copie a chave pública: `cat ~/.ssh/sua_chave.pub | pbcopy`
2. Adicione em: https://github.com/settings/keys

### Problema: Chave errada está sendo usada

**Causa:** Múltiplas chaves carregadas no ssh-agent

**Solução:**

```bash
# Limpar todas as chaves
ssh-add -D

# Adicionar apenas as chaves necessárias
ssh-add ~/.ssh/id_ed25519
ssh-add ~/.ssh/id_ed25519_personal

# Ou especificar no SSH config com IdentitiesOnly yes
```

## 📊 Resumo da Configuração

### Arquivos Criados/Modificados

```
~/.ssh/config                    # Configuração SSH
~/.ssh/id_ed25519               # Chave privada corporativa
~/.ssh/id_ed25519.pub           # Chave pública corporativa
~/.ssh/id_ed25519_personal      # Chave privada pessoal
~/.ssh/id_ed25519_personal.pub  # Chave pública pessoal
~/.gitconfig                    # Configuração Git principal
~/.gitconfig-empresax           # Configuração Git corporativa
```

### Fluxo de Decisão

```
┌─────────────────────────┐
│  Clone/Push Repository  │
└───────────┬─────────────┘
            │
            ▼
    ┌───────────────┐
    │ Check Remote  │
    │     URL       │
    └───────┬───────┘
            │
     ┌──────┴──────┐
     │             │
     ▼             ▼
┌─────────┐   ┌─────────┐
│ Contains│   │  Other  │
│empresax │   │  repos  │
└────┬────┘   └────┬────┘
     │             │
     ▼             ▼
┌─────────┐   ┌─────────┐
│Corporate│   │Personal │
│  Config │   │  Config │
│         │   │         │
│ email@  │   │email@   │
│empresa  │   │gmail    │
│         │   │         │
│id_ed    │   │id_ed25  │
│25519    │   │519_pers │
└─────────┘   └─────────┘
```

## 📚 Referências

- [Git Config Documentation](https://git-scm.com/docs/git-config)
- [SSH Config Documentation](https://man.openbsd.org/ssh_config)
- [GitHub: Multiple SSH Keys](https://docs.github.com/en/authentication/connecting-to-github-with-ssh)
- [Git Conditional Includes](https://git-scm.com/docs/git-config#_conditional_includes)

## 💡 Dicas Extras

### 1. Verificar configuração antes de commitar

Crie um alias Git para verificar rapidamente qual conta está sendo usada:

```bash
git config --global alias.whoami '!echo "Email: $(git config user.email)" && echo "Key: $(git config user.signingkey)"'
```

Use antes de commitar:

```bash
git whoami
```

### 2. Aliases úteis no shell

Adicione ao `~/.zshrc` ou `~/.bashrc`:

```bash
# Aliases SSH
alias sshempresax="ssh-add ~/.ssh/id_ed25519"
alias sshpessoal="ssh-add ~/.ssh/id_ed25519_personal"
alias sshlist="ssh-add -l"
alias sshclear="ssh-add -D"

# Alias Git
alias gitinfo="git config user.email && git config user.signingkey"
alias gitkeys="ssh-add -l | grep -E 'gmail|empresa'"
```

### 3. Assinatura automática de commits

A configuração `commit.gpgsign = true` garante que todos os commits sejam assinados automaticamente com a chave SSH configurada.

Para verificar se um commit está assinado:

```bash
git log --show-signature -1
```

## ✅ Checklist Final

Antes de começar a usar, verifique se tudo está configurado:

- [ ] Duas chaves SSH criadas (`id_ed25519` e `id_ed25519_personal`)
- [ ] Chaves adicionadas às respectivas contas do GitHub
- [ ] Arquivo `~/.ssh/config` configurado
- [ ] Arquivo `~/.gitconfig-empresax` criado
- [ ] Arquivo `~/.gitconfig` atualizado com `includeIf`
- [ ] Chaves adicionadas ao ssh-agent
- [ ] Testado conexão SSH: `ssh -T git@github.com` e `ssh -T git@github.com-empresax`
- [ ] Testado em repositório corporativo: `git config user.email`
- [ ] Testado em repositório pessoal: `git config user.email`
- [ ] Aliases criados no `~/.zshrc` (opcional)

---

**Pronto!** 🎉 Agora você tem uma configuração profissional que gerencia automaticamente múltiplas contas Git!

## 🤝 Contribuindo

Encontrou algum erro ou tem sugestões? Sinta-se à vontade para contribuir!