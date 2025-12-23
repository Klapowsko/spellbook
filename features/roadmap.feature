Feature: Gerar Roadmap de Estudo
  Como um cliente do serviço Spellbook
  Eu quero gerar um roadmap de estudo
  Para que eu possa ter um guia estruturado de aprendizado

  Scenario: Gerar roadmap com sucesso
    Given que tenho uma API key válida do Gemini
    When eu envio uma requisição POST para /roadmap com topic "Machine Learning"
    Then a resposta deve ter status 200
    And a resposta deve conter um roadmap com pelo menos 4 categorias
    And cada categoria deve ter entre 5 e 10 itens
    And o roadmap deve ter o topic "Machine Learning"

  Scenario: Gerar roadmap com tópico vazio
    Given que tenho uma API key válida do Gemini
    When eu envio uma requisição POST para /roadmap com topic ""
    Then a resposta deve ter status 400

  Scenario: Gerar roadmap sem API key
    Given que não tenho uma API key do Gemini configurada
    When eu envio uma requisição POST para /roadmap com topic "Python"
    Then a resposta deve ter status 500
    And a resposta deve conter uma mensagem de erro sobre API key

