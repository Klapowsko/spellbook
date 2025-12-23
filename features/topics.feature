Feature: Gerar Tópicos sobre um Assunto
  Como um cliente do serviço Spellbook
  Eu quero gerar uma lista de tópicos sobre um assunto
  Para que eu possa explorar diferentes áreas de conhecimento

  Scenario: Gerar lista de tópicos com sucesso
    Given que tenho uma API key válida do Gemini
    When eu envio uma requisição POST para /topics com subject "Python" e count 10
    Then a resposta deve ter status 200
    And a resposta deve conter uma lista de tópicos
    And a lista deve ter pelo menos 5 tópicos
    And o subject deve ser "Python"

  Scenario: Gerar tópicos sem especificar count
    Given que tenho uma API key válida do Gemini
    When eu envio uma requisição POST para /topics com subject "JavaScript"
    Then a resposta deve ter status 200
    And a resposta deve conter uma lista de tópicos
    And a lista deve ter pelo menos 5 tópicos

  Scenario: Gerar tópicos com subject vazio
    Given que tenho uma API key válida do Gemini
    When eu envio uma requisição POST para /topics com subject ""
    Then a resposta deve ter status 400

